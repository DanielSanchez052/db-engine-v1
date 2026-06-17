****# GO DB Engine V1 - Roadmap de Implementación

## Objetivo

Construir un motor de base de datos relacional embebido en Go, con dependencias mínimas, inspirado en arquitecturas reales como SQLite.

---

# Decisiones Arquitectónicas

## Motor

* Embedded Database
* Sin cliente-servidor
* Sin sockets
* Sin protocolo de red

## Lenguaje

* Go

## Persistencia

* `database.db`
* `catalog.db`

## Páginas

* Tamaño fijo: 4096 bytes
* Header: 64 bytes
* Payload: 4032 bytes

## Organización de Archivos

```text
database.db

Page 0  -> Database Header
Page N  -> Data Pages
```

## Layout de Registros

Slotted Pages

```text
┌────────────────────┐
│ Header             │
├────────────────────┤
│ Slots              │
├────────────────────┤
│ Free Space         │
├────────────────────┤
│ Records            │
└────────────────────┘
```

## Catálogo

Archivo independiente:

```text
catalog.db
```

---

# Fase 1 - Fundaciones del Storage Engine

Objetivo: Poder crear y abrir una base de datos.

## 1. Database Header

### Componentes

* [x] Magic Number
* [x] Version
* [x] PageSize
* [x] TotalPages
* [x] FreePageHead

### Archivos

```text
internal/storage/database/
├── header.go
├── constants.go
└── errors.go
```

### Resultado esperado

* Crear una base de datos nueva.
* Escribir Page 0.
* Leer Page 0.

---

## 2. FileManager

### Responsabilidades

* Abrir archivo
* Leer bytes
* Escribir bytes
* Flush a disco

### Archivos

```text
internal/storage/filemanager/
├── file_manager.go
└── errors.go
```

### Funcionalidades

* [x] Open
* [x] Close
* [x] ReadAt
* [x] WriteAt
* [x] Sync

### Resultado esperado

Lectura y escritura arbitraria mediante offsets.

---

## 3. Page

### Responsabilidades

Representación en memoria de una página.

### Archivos

```text
internal/storage/page/
├── page.go
├── page_header.go
├── page_type.go
└── constants.go
```

### Funcionalidades

* [X] Crear página vacía
* [X] Serializar página
* [X] Deserializar página

### Resultado esperado

Transformar:

```text
Page -> []byte
```

y

```text
[]byte -> Page
```

---

## 4. Slot

### Responsabilidades

Representación de una entrada del Slot Directory.

### Archivos

```text
internal/storage/slot/
├── slot.go
└── constants.go
```

### Funcionalidades

* [ ] Crear Slot
* [ ] Serializar Slot
* [ ] Deserializar Slot
* [ ] Detectar Slot eliminado

### Resultado esperado

Administrar registros mediante referencias indirectas.

---

## 5. Pager

### Responsabilidades

Administrar páginas dentro del archivo.

### Archivos

```text
internal/storage/pager/
├── pager.go
├── allocator.go
└── errors.go
```

### Funcionalidades

* [ ] AllocatePage
* [ ] GetPage
* [ ] FlushPage

### Resultado esperado

Administración básica de páginas persistentes.

---

# Fase 2 - Slotted Pages

Objetivo: Almacenar registros dentro de páginas.

## Funcionalidades

* [ ] InsertRecord
* [ ] GetRecord
* [ ] DeleteRecord
* [ ] Reutilización de slots libres

### Resultado esperado

Poder almacenar registros binarios arbitrarios.

---

# Fase 3 - Catálogo

Objetivo: Definir tablas y esquemas.

## Componentes

### Table

* [ ] Nombre
* [ ] Id

### Column

* [ ] Nombre
* [ ] Tipo
* [ ] Nullable

### Schema

* [ ] Colección de columnas

### Archivos

```text
internal/catalog/
├── catalog.go
├── table.go
├── column.go
└── schema.go
```

### Resultado esperado

Registrar y recuperar definiciones de tablas.

---

# Fase 4 - Query Layer

Objetivo: Crear una API mínima para operar tablas.

## Funcionalidades

* [ ] CreateTable
* [ ] Insert
* [ ] SelectAll

### Archivos

```text
internal/executor/
├── create_table.go
├── insert.go
└── select.go
```

### Resultado esperado

Ejecutar operaciones básicas sobre tablas.

---

# Fase 5 - Parser SQL

Objetivo: Interpretar SQL básico.

## Funcionalidades

* [ ] CREATE TABLE
* [ ] INSERT INTO
* [ ] SELECT *

### Archivos

```text
internal/parser/
└── parser.go
```

### Resultado esperado

Convertir SQL en operaciones ejecutables.

---

# Fase 6 - MVP Completo

## Debe permitir

```sql
CREATE TABLE users (
    id INT,
    name STRING
);

INSERT INTO users VALUES (1, 'Daniel');

SELECT * FROM users;
```

## Fuera de alcance

* Índices
* JOINs
* Foreign Keys
* Transactions
* WAL
* MVCC
* Query Optimizer
* Cliente/Servidor

---

# Fase 7 - Evolución Futura

## Storage

* [ ] Free Page List
* [ ] Buffer Pool

## Índices

* [ ] B+Tree
* [ ] Índices primarios

## SQL

* [ ] WHERE
* [ ] ORDER BY
* [ ] LIMIT

## Integridad

* [ ] Primary Keys
* [ ] Foreign Keys

## Concurrencia

* [ ] Locks
* [ ] Transactions

## Recuperación

* [ ] Write Ahead Log (WAL)
* [ ] Crash Recovery

```
```
****