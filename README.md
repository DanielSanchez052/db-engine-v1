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
- [x] Slot
- [x] Pager
- [x] Slotted Pages
- [x] HeapFile
- [x] Database

### Pendiente

- [ ] Catalog
- [ ] Executor
- [ ] Parser SQL

### Future Work:
- [ ] Free Space Map
- [ ] Page Directory

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

# Arquitectura Decidida

Esta sección documenta las decisiones arquitectónicas que forman parte del contrato fundamental de MiniDB.

Las decisiones aquí registradas se consideran estables y cualquier modificación futura debe evaluarse cuidadosamente debido a su impacto sobre la compatibilidad de los archivos persistidos.

---

## Tipo de Motor

MiniDB es un motor:

```text
Embedded Database
```

Características:

- Sin servidor
- Sin sockets
- Sin protocolo de red
- Integración directa como librería
- Acceso directo a archivos

Inspiración:

- SQLite

---

## Lenguaje

```text
Go
```

Motivaciones:

- Simplicidad
- Excelente soporte para sistemas
- Manejo eficiente de archivos
- Binarios estáticos
- Biblioteca estándar robusta

---

## Dependencias

Objetivo:

```text
Standard Library Only
```

MiniDB intentará minimizar dependencias externas para comprender e implementar cada componente internamente.

---

## Persistencia

La persistencia se realiza mediante archivos separados.

```text
database.db
catalog.db
```

### database.db

Contiene:

- Database Header
- Data Pages
- Index Pages (futuro)

### catalog.db

Contiene:

- Tablas
- Columnas
- Metadatos

---

## Tamaño de Página

Todas las páginas utilizan tamaño fijo.

```text
4096 bytes
```

Constante:

```go
const PageSize = 4096
```

Motivaciones:

- Simplicidad
- Alineación con sistemas reales
- Fácil administración de offsets

---

## Layout General de Página

```text
┌────────────────────┐
│ Page Header        │ 64 bytes
├────────────────────┤
│ Payload            │ 4032 bytes
└────────────────────┘
```

---

## Database Header

La página cero está reservada para metadatos globales.

```text
Page 0
```

Layout:

```text
Offset  Size  Field

0       4     Magic Number
4       2     Version
6       2     Page Size
8       8     Total Pages
16      8     Free Page Head
24      40    Reserved
```

Tamaño:

```text
64 bytes
```

---

## Page Header

Cada página contiene un encabezado propio.

Layout:

```text
Offset  Size  Field

0       8     Page ID
8       2     Record Count
10      2     Free Space Offset
12      2     Slot Count
14      1     Page Type
15      49    Reserved
```

Tamaño:

```text
64 bytes
```

---

## Endianness

Todo el formato binario utiliza:

```text
Little Endian
```

MiniDB no depende del layout interno de los structs de Go.

Toda serialización se realiza mediante offsets explícitos.

---

## Formato de Registros

Los registros serán almacenados como:

```go
type Record []byte
```

El Storage Engine no conoce:

- Tablas
- Columnas
- Tipos

Solo administra bytes.

La interpretación de los registros será responsabilidad de capas superiores.

---

## Layout Interno de Páginas

MiniDB utilizará:

```text
Slotted Pages
```

Layout objetivo:

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

Motivaciones:

- Inserciones eficientes
- Eliminaciones lógicas
- Reutilización de espacio
- Independencia entre ubicación física y referencias

---

## Eliminación de Registros

Un slot eliminado se representará mediante:

```text
RecordOffset = 0
RecordLength = 0
```

Esto permitirá reutilizar slots sin reorganizar el directorio completo.

---

## Propiedad de Archivos

El acceso a disco se realiza mediante:

```go
type FileManager struct {
    file *os.File
}
```

FileManager es propietario del descriptor de archivo.

No se utilizan interfaces de abstracción en la V1.

---

## Alcance de la V1

Incluido:

- Storage Engine
- Catálogo
- CREATE TABLE
- INSERT
- SELECT *

Excluido:

- Índices
- JOIN
- Foreign Keys
- WAL
- MVCC
- Query Optimizer
- Concurrencia
- Cliente/Servidor

---

## Compatibilidad

Una vez existan bases de datos persistidas, las siguientes características se consideran parte del formato estable:

- Page Size
- Database Header
- Page Header
- Endianness
- Slotted Pages
- Magic Number

Cualquier cambio deberá tratarse como una nueva versión del formato de almacenamiento.

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