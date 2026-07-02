package migrator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"go.yaml.in/yaml/v4"
)

// Migrator is the core engine for schema migration
type Migrator struct {
	Client *lark.Client
}

// NewMigrator initializes a Lark Bitable migrator
func NewMigrator(appID, appSecret string) *Migrator {
	// Create Lark official client
	client := lark.NewClient(appID, appSecret)
	return &Migrator{
		Client: client,
	}
}

// LoadConfig reads the bitable.yaml file
func (m *Migrator) LoadConfig(path string) (*BitableConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config BitableConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveConfig writes the BitableConfig back to the yaml file atomically
func (m *Migrator) SaveConfig(path string, config *BitableConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	tempFile := path + ".tmp"
	err = ioutil.WriteFile(tempFile, data, 0644)
	if err != nil {
		return err
	}

	err = os.Rename(tempFile, path)
	if err != nil {
		os.Remove(tempFile)
		return err
	}
	return nil
}

// FetchOnlineFields retrieves the list of fields for a table on Feishu (handling pagination)
func (m *Migrator) FetchOnlineFields(appToken, tableID string) ([]*larkbitable.AppTableFieldForList, error) {
	var allFields []*larkbitable.AppTableFieldForList
	var pageToken string
	pageSize := 100 // Request maximum of 100 items per page to reduce API calls

	for {
		reqBuilder := larkbitable.NewListAppTableFieldReqBuilder().
			AppToken(appToken).
			TableId(tableID).
			PageSize(pageSize)

		if pageToken != "" {
			reqBuilder.PageToken(pageToken)
		}

		req := reqBuilder.Build()
		resp, err := m.Client.Bitable.AppTableField.List(context.Background(), req)
		if err != nil {
			return nil, err
		}
		if !resp.Success() {
			return nil, fmt.Errorf("lark api error: code=%d, msg=%s", resp.Code, resp.Msg)
		}

		if resp.Data != nil && resp.Data.Items != nil {
			allFields = append(allFields, resp.Data.Items...)
		}

		// Exit loop if no more pages are available
		if resp.Data == nil || resp.Data.HasMore == nil || !*resp.Data.HasMore || resp.Data.PageToken == nil || *resp.Data.PageToken == "" {
			break
		}
		pageToken = *resp.Data.PageToken
	}

	return allFields, nil
}

// toLarkProperty converts dynamic property maps to larkbitable.AppTableFieldProperty
func toLarkProperty(propMap map[string]interface{}) *larkbitable.AppTableFieldProperty {
	if len(propMap) == 0 {
		return nil
	}
	jsonBytes, err := json.Marshal(propMap)
	if err != nil {
		return nil
	}
	var prop larkbitable.AppTableFieldProperty
	err = json.Unmarshal(jsonBytes, &prop)
	if err != nil {
		return nil
	}
	return &prop
}

// getDescriptionText extracts description string from standard interface{} returned by Lark SDK
func getDescriptionText(desc interface{}) string {
	if desc == nil {
		return ""
	}
	if m, ok := desc.(map[string]interface{}); ok {
		if text, ok := m["text"].(string); ok {
			return text
		}
	}
	jsonBytes, err := json.Marshal(desc)
	if err == nil {
		var m map[string]interface{}
		if json.Unmarshal(jsonBytes, &m) == nil {
			if text, ok := m["text"].(string); ok {
				return text
			}
		}
	}
	return ""
}

// propertiesEqual compares two properties using JSON serialization comparison
func propertiesEqual(p1 map[string]interface{}, p2 *larkbitable.AppTableFieldProperty) bool {
	if len(p1) == 0 && p2 == nil {
		return true
	}
	// Dynamic unmarshal of p2 to compare structure
	jsonBytes2, err := json.Marshal(p2)
	if err != nil {
		return false
	}
	var m2 map[string]interface{}
	json.Unmarshal(jsonBytes2, &m2)

	// Clean up empty fields from both to make comparison fair
	cleanMap := func(m map[string]interface{}) map[string]interface{} {
		res := make(map[string]interface{})
		for k, v := range m {
			if v == nil {
				continue
			}
			// Skip nil/empty options or maps
			if rv := reflect.ValueOf(v); rv.Kind() == reflect.Slice && rv.Len() == 0 {
				continue
			}
			res[k] = v
		}
		return res
	}

	m1Clean := cleanMap(p1)
	m2Clean := cleanMap(m2)

	return reflect.DeepEqual(m1Clean, m2Clean)
}

// Sync performs the two-way diff comparison and applies changes
func (m *Migrator) Sync(path string, dryRun bool) error {
	config, err := m.LoadConfig(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if config.AppToken == "" {
		return fmt.Errorf("app_token must be provided in config")
	}

	configUpdated := false

	for tIdx, table := range config.Tables {
		if table.TableID == "" {
			log.Printf("[Warning] Table %s is missing table_id, skipping.", table.Name)
			continue
		}

		log.Printf("[Info] Fetching fields for table '%s' (%s)...", table.Name, table.TableID)
		onlineFields, err := m.FetchOnlineFields(config.AppToken, table.TableID)
		if err != nil {
			return fmt.Errorf("failed to fetch fields for table %s: %w", table.Name, err)
		}

		// Create quick lookup maps for online fields
		onlineByID := make(map[string]*larkbitable.AppTableFieldForList)
		onlineByName := make(map[string]*larkbitable.AppTableFieldForList)
		for _, f := range onlineFields {
			if f.FieldId != nil {
				onlineByID[*f.FieldId] = f
			}
			if f.FieldName != nil {
				onlineByName[*f.FieldName] = f
			}
		}

		for fIdx, field := range table.Fields {
			var matchedOnline *larkbitable.AppTableFieldForList
			matchedByID := false

			// Match by ID first, then by name
			if field.FieldID != "" {
				if f, ok := onlineByID[field.FieldID]; ok {
					matchedOnline = f
					matchedByID = true
				}
			} else if f, ok := onlineByName[field.Name]; ok {
				matchedOnline = f
			}

			if matchedOnline == nil {
				// 1. Create Field
				log.Printf("[Create] Field '%s' (Type %s) does not exist online. Creating...", field.Name, field.Type)
				if !dryRun {
					prop := toLarkProperty(field.Property)
					builder := larkbitable.NewAppTableFieldBuilder().
						FieldName(field.Name).
						Type(int(field.Type)).
						Property(prop).
						Description(larkbitable.NewAppTableFieldDescriptionBuilder().
							Text(field.Description).
							Build())
					if field.UiType != "" {
						builder.UiType(field.UiType)
					}
					reqField := builder.Build()

					req := larkbitable.NewCreateAppTableFieldReqBuilder().
						AppToken(config.AppToken).
						TableId(table.TableID).
						AppTableField(reqField).
						Build()

					resp, err := m.Client.Bitable.AppTableField.Create(context.Background(), req)
					if err != nil {
						return fmt.Errorf("failed to create field %s: %w", field.Name, err)
					}
					if !resp.Success() {
						return fmt.Errorf("failed to create field %s: code=%d, msg=%s", field.Name, resp.Code, resp.Msg)
					}

					newID := *resp.Data.Field.FieldId
					log.Printf("[Success] Field '%s' created successfully with ID: %s", field.Name, newID)

					// Update local config memory
					config.Tables[tIdx].Fields[fIdx].FieldID = newID
					configUpdated = true
				}
			} else {
				// If matched by name but ID is missing locally, bind the ID!
				if !matchedByID {
					log.Printf("[Bind] Binding existing field '%s' online ID '%s' to local YAML.", field.Name, *matchedOnline.FieldId)
					config.Tables[tIdx].Fields[fIdx].FieldID = *matchedOnline.FieldId
					configUpdated = true
					field.FieldID = *matchedOnline.FieldId
				}

				// 2. Check for updates
				needsUpdate := false
				onlineName := ""
				if matchedOnline.FieldName != nil {
					onlineName = *matchedOnline.FieldName
				}
				onlineType := 0
				if matchedOnline.Type != nil {
					onlineType = *matchedOnline.Type
				}
				onlineDesc := getDescriptionText(matchedOnline.Description)
				onlineUiType := ""
				if matchedOnline.UiType != nil {
					onlineUiType = *matchedOnline.UiType
				}

				if onlineName != field.Name {
					log.Printf("[Update] Field '%s' will be renamed online to '%s'", onlineName, field.Name)
					needsUpdate = true
				}
				if field.UiType != "" && onlineUiType != field.UiType {
					log.Printf("[Update] Field '%s' ui_type will change from '%s' to '%s'", field.Name, onlineUiType, field.UiType)
					needsUpdate = true
				}
				if onlineType != int(field.Type) {
					log.Printf("[Update] Field '%s' type will change from %s to %s", field.Name, FieldType(onlineType), field.Type)
					needsUpdate = true
				}
				if onlineDesc != field.Description {
					log.Printf("[Update] Field '%s' description will change from '%s' to '%s'", field.Name, onlineDesc, field.Description)
					needsUpdate = true
				}
				if !propertiesEqual(field.Property, matchedOnline.Property) {
					log.Printf("[Update] Field '%s' properties will be updated.", field.Name)
					needsUpdate = true
				}

				if needsUpdate {
					log.Printf("[Update] Field '%s' (%s) properties differ. Updating...", field.Name, field.FieldID)
					if !dryRun {
						prop := toLarkProperty(field.Property)
						builder := larkbitable.NewAppTableFieldBuilder().
							FieldName(field.Name).
							Type(int(field.Type)).
							Property(prop).
							Description(larkbitable.NewAppTableFieldDescriptionBuilder().
								Text(field.Description).
								Build())
						if field.UiType != "" {
							builder.UiType(field.UiType)
						}
						reqField := builder.Build()

						req := larkbitable.NewUpdateAppTableFieldReqBuilder().
							AppToken(config.AppToken).
							TableId(table.TableID).
							FieldId(field.FieldID).
							AppTableField(reqField).
							Build()

						resp, err := m.Client.Bitable.AppTableField.Update(context.Background(), req)
						if err != nil {
							return fmt.Errorf("failed to update field %s (%s): %w", field.Name, field.FieldID, err)
						}
						if !resp.Success() {
							return fmt.Errorf("failed to update field %s (%s): code=%d, msg=%s", field.Name, field.FieldID, resp.Code, resp.Msg)
						}
						log.Printf("[Success] Field '%s' (%s) updated successfully.", field.Name, field.FieldID)
					}
				} else {
					if !dryRun {
						log.Printf("[No Change] Field '%s' (%s) is up to date.", field.Name, field.FieldID)
					}
				}
			}
		}
	}

	// Automatical writeback to yaml file if changes occurred
	if configUpdated && !dryRun {
		log.Printf("[Save] Writing back newly assigned field IDs to '%s'...", path)
		err = m.SaveConfig(path, config)
		if err != nil {
			return fmt.Errorf("failed to write config back to yaml: %w", err)
		}
		log.Printf("[Success] Configuration saved successfully.")
	}

	return nil
}

// Pull generates a new yaml schema from the online multi-dimensional table
	func (m *Migrator) Pull(appToken, tableID, outputPath string) error {
	log.Printf("[Pull] Fetching schema for app: %s, table: %s...", appToken, tableID)
	fields, err := m.FetchOnlineFields(appToken, tableID)
	if err != nil {
		return err
	}

	log.Printf("[Pull] Retrieved %d fields from Bitable.", len(fields))

	var localFields []FieldConfig
	for _, f := range fields {
		var fieldConf FieldConfig
		if f.FieldId != nil {
			fieldConf.FieldID = *f.FieldId
		}
		if f.FieldName != nil {
			fieldConf.Name = *f.FieldName
		}
		if f.Type != nil {
			fieldConf.Type = FieldType(*f.Type)
		}
		if f.UiType != nil {
			fieldConf.UiType = *f.UiType
		}
		fieldConf.Description = getDescriptionText(f.Description)

		if f.Property != nil {
			// Convert Property to map for clean yaml export
			jsonBytes, err := json.Marshal(f.Property)
			if err == nil {
				var propMap map[string]interface{}
				if json.Unmarshal(jsonBytes, &propMap) == nil {
					// Remove null or empty elements
					for k, v := range propMap {
						if v == nil {
							delete(propMap, k)
						}
					}
					if len(propMap) > 0 {
						fieldConf.Property = propMap
					}
				}
			}
		}

		localFields = append(localFields, fieldConf)
	}

	config := &BitableConfig{
		AppToken: appToken,
		Tables: []TableConfig{
			{
				TableID: tableID,
				Name:    "Pulled Table",
				Fields:  localFields,
			},
		},
	}

	// Create parent dirs if necessary
	dir := filepath.Dir(outputPath)
	if dir != "." {
		os.MkdirAll(dir, 0755)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outputPath, data, 0644)
	if err != nil {
		return err
	}

	log.Printf("[Success] Schema successfully generated and saved to '%s'", outputPath)
	return nil
}
