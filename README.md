# Bitable Migrator (飞书多维表格声明式 Schema 迁移与同步工具)

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)

**Bitable Migrator** is a lightweight, declarative, and high-performance schema migration tool for **Lark/Feishu Bitable (多维表格)** written in Go. 

It allows developer teams to manage Bitable tables and fields as code (**Schema-as-Code**), keeping Bitable structures inside Git, synchronizing changes safely, and coexisting perfectly with other manual edits in multi-user collaborative environments.

---

## 🌟 Key Features

1. **Declarative YAML Schema**: Define your Bitable App Token, Tables, and Fields in a simple `bitable.yaml` file. No sequential migration files or SQL scripts to write!
2. **Automatic `field_id` Writeback**: Newly created or matched field IDs are automatically written back and saved into your local `bitable.yaml` file upon synchronization. Perfect for Git tracking!
3. **Rename & Restructure Safe**: Since field IDs are securely tracked in the YAML file, renaming fields or updating options will always target the correct field via the Lark API, preventing duplicate column creation.
4. **Graceful Coexistence (Zero Interference)**: Bitable Migrator strictly adheres to the "maintain only our own fields" rule. Any online fields not declared in the YAML config are ignored and preserved.
5. **Dry-Run Diff Check**: Preview what changes (create, update, bind) will be applied to Bitable before actually making the network calls.
6. **Reverse Schema Generation (Pull)**: Generate a complete `bitable.yaml` instantly from an existing Bitable table.

---

## 🚀 Installation & Build

Ensure you have Go 1.26+ installed. Clone or copy this repository to your Go workspace, then run:

```bash
cd bitable-migrator
go mod tidy
go build -o bin/bitable-migrator .
```

This compiles a standalone, high-performance `bitable-migrator` binary under the `bin/` directory.

---

## 📁 JSON Schema Autocomplete & Validation

To enable rich autocomplete prompts, hover tooltips, and validation in VS Code, JetBrains IDEs, or any editor supporting YAML language server, add this comment at the very first line of your `bitable.yaml`:

```yaml
# yaml-language-server: $schema=./bitable.schema.json

app_token: "appbcbWCzen6D8dezhoCH2RpMAh"
tables:
  - table_id: "tblsRc9GRRXKqhvW"
    name: "User Table"
    fields:
      - name: "Full Name"
        type: 1 # autocomplete will show this is Text
```

---

## 🛠️ Usage & Commands

Export your Lark App credentials before running:

```bash
export LARK_APP_ID="cli_xxxxxxxxxxx"
export LARK_APP_SECRET="xxxxxxxxxxxxxxxxxxxxx"
```

### 1. Synchronize Changes (`sync`)
Compares the local `bitable.yaml` with the online schema, applies updates/creations, and writes back newly allocated `field_id`s back into the local YAML file.

```bash
./bin/bitable-migrator sync --config ./bitable.yaml
```

### 2. Preview Changes (`diff`)
Dry-run mode. Compares the structures and prints the planned creations or modifications without modifying anything online or locally.

```bash
./bin/bitable-migrator diff --config ./bitable.yaml
```

### 3. Generate Schema from Online Table (`pull`)
Initializes a new `bitable.yaml` base template from an existing Feishu Bitable.

```bash
./bin/bitable-migrator pull --app <app_token> --table <table_id> --output ./bitable.yaml
```

---

## ⚙️ Supported Field Types Map

Below are Bitable's standard field type IDs supported inside the `type` property:

| Type ID | Type Name | Detail |
| :--- | :--- | :--- |
| **1** | Text | Multi-line text field |
| **2** | Number | Numeric field |
| **3** | SingleSelect | Dropdown with unique selection |
| **4** | MultiSelect | Dropdown with multiple tags |
| **5** | Date | Calendar Date |
| **7** | Checkbox | Boolean checkbox |
| **11** | User | Member/Personnel selector |
| **18** | SingleLink | Reference to another table |
| **20** | Formula | Mathematical or text formula expression |
| **21** | DuplexLink | Two-way reference connection |
| **1001** | CreatedTime | Metadata: Record creation timestamp |
| **1002** | ModifiedTime | Metadata: Last modification timestamp |

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
