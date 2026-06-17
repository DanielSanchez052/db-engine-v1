# GO DB Engine V1

GO DB Engine V1 es un motor de base de datos relacional embebido escrito en Go, construido desde cero con fines educativos y de aprendizaje sobre arquitectura interna de sistemas de almacenamiento.

El objetivo del proyecto es comprender e implementar los componentes fundamentales de una base de datos moderna:

- Gestión de archivos
- Gestión de páginas
- Slotted Pages
- Catálogo de metadatos
- Parser SQL
- Ejecutor de consultas
- Índices
- Recuperación y concurrencia (futuras versiones)

---

# Objetivos

- Comprender cómo funciona internamente un motor de base de datos.
- Minimizar dependencias externas.
- Implementar formatos binarios explícitos.
- Mantener una arquitectura simple y extensible.
- Evolucionar progresivamente desde un almacenamiento básico hasta un motor relacional funcional.

---

# Estado Actual

## Implementado

### Storage Layer

- [x] Database Header
- [x] File Manager
- [x] Page Header
- [x] Page
- [x] Serialización binaria
- [x] Deserialización binaria
- [x] Tests unitarios

### Pendiente

- [ ] Slot
- [ ] Pager
- [ ] Slotted Pages
- [ ] Catalog
- [ ] Executor
- [ ] Parser SQL

---

# Arquitectura

```text
Engine
 │
 ▼
Catalog
 │
 ▼
Executor
 │
 ▼
Pager
 │
 ▼
Page
 │
 ▼
FileManager
 │
 ▼
database.db
```

---

# Filosofía de Diseño

GO DB Engine V1 adopta una filosofía similar a SQLite:

- Embedded Database
- Sin servidor
- Sin sockets
- Sin protocolo de red
- Acceso directo a archivos

La base de datos es consumida como una librería dentro de una aplicación.

```go
db, err := gedb.Open("database.db")
```

---

# Estructura del Proyecto

```text
gedb/
│
├── cmd/
│   └── gedb/
│
├── internal/
│   │
│   ├── storage/
│   │   ├── database/
│   │   ├── filemanager/
│   │   ├── page/
│   │   ├── slot/
│   │   └── pager/
│   │
│   ├── catalog/
│   ├── executor/
│   ├── parser/
│   └── engine/
│
├── test/
│
└── data/
```

---

# Persistencia

GO DB Engine V1 utiliza dos archivos físicos.

## database.db

Contiene:

- Header de base de datos
- Páginas de datos
- Páginas de índices (futuro)

## catalog.db

Contiene:

- Definiciones de tablas
- Columnas
- Metadatos del sistema

---

# Formato de Página

Todas las páginas tienen tamaño fijo.

```text
Page Size = 4096 bytes
```

---

## Distribución

```text
┌────────────────────┐
│ Page Header        │ 64 bytes
├────────────────────┤
│ Payload            │ 4032 bytes
└────────────────────┘

Total = 4096 bytes
```

---

# Database Header

La página 0 está reservada para información global de la base de datos.

```text
Offset  Size  Field

0       4     Magic Number
4       2     Version
6       2     Page Size
8       8     Total Pages
16      8     Free Page Head
24      40    Reserved
```

Tamaño total:

```text
64 bytes
```

---

# Page Header

Cada página contiene un encabezado propio.

```text
Offset  Size  Field

0       8     Page ID
8       2     Record Count
10      2     Free Space Offset
12      2     Slot Count
14      1     Page Type
15      49    Reserved
```

Tamaño total:

```text
64 bytes
```

---

# Tipos de Página

```go
const (
    DataPage PageType = iota
    IndexPage
    CatalogPage
)
```

---

# Formato Binario

Todos los formatos binarios utilizan:

```text
Little Endian
```

y offsets explícitos.

GO DB Engine V1 no depende del layout interno de los structs de Go.

---

# Layout de Registros (Objetivo)

Las páginas de datos utilizarán el patrón:

```text
Slotted Page
```

```text
┌────────────────────┐
│ Header             │
├────────────────────┤
│ Slot Directory     │
├────────────────────┤
│ Free Space         │
├────────────────────┤
│ Records            │
└────────────────────┘
```

Esto permitirá:

- Inserciones eficientes
- Eliminaciones lógicas
- Reutilización de espacio
- Movimiento de registros sin invalidar referencias

---

# Roadmap

## Fase 1 - Fundaciones

- [x] Database Header
- [x] File Manager
- [x] Page Header
- [x] Page

## Fase 2 - Storage Engine

- [ ] Slot
- [ ] Pager
- [ ] Record Management
- [ ] Slotted Pages

## Fase 3 - Catálogo

- [ ] Tables
- [ ] Columns
- [ ] Schemas

## Fase 4 - Query Engine

- [ ] CREATE TABLE
- [ ] INSERT
- [ ] SELECT

## Fase 5 - SQL Parser

- [ ] CREATE TABLE
- [ ] INSERT INTO
- [ ] SELECT *

## Fase 6 - MVP

Objetivo final de la primera versión:

```sql
CREATE TABLE users (
    id INT,
    name STRING
);

INSERT INTO users VALUES (1, 'Daniel');

SELECT * FROM users;
```

---

# Futuras Versiones

## Storage

- Free Page List
- Buffer Pool

## Índices

- B+Tree
- Índices primarios

## SQL

- WHERE
- ORDER BY
- LIMIT

## Integridad

- Primary Keys
- Foreign Keys

## Recuperación

- WAL
- Crash Recovery

## Concurrencia

- Locks
- Transactions

## Distribución

- Cliente/Servidor (opcional)

---

# Licencia

Proyecto educativo y experimental.