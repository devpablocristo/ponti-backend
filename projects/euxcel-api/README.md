# Euxcel API

Este proyecto está diseñado para ejecutarse en entornos de desarrollo y staging usando Docker Compose. A continuación, se detallan los posibles problemas, soluciones y comandos disponibles para compilar, ejecutar y administrar el proyecto.

---

## 🛠️ **Posible problema con MongoDB y solución**
Si tienes problemas para inicializar MongoDB y crear usuarios automáticamente con un script de inicialización (`init-mongo.js`), puedes seguir los siguientes pasos para ejecutarlo manualmente:

1. Abre una terminal y accede al contenedor de MongoDB (suponiendo que el contenedor se llame `mongodb`):

```bash
docker exec -it mongodb mongo --username root --password rootpassword --authenticationDatabase admin
```

2. Una vez dentro de la shell de MongoDB, carga y ejecuta el script de inicialización con el comando `load`. Si el script se encuentra en `/docker-entrypoint-initdb.d/init-mongo.js`, puedes ejecutarlo así:

```js
load("/docker-entrypoint-initdb.d/init-mongo.js")
```

Esto ejecutará el script y creará el usuario en la base de datos especificada. Asegúrate de que la ruta del script sea correcta según dónde se encuentre dentro del contenedor.

---

## 🚀 **Comandos disponibles**

El archivo `Makefile` permite ejecutar rápidamente tareas de compilación, ejecución y pruebas mediante Docker Compose y Go.

---

### ✅ **Comandos básicos**
| Comando         | Descripción                                                                                           |
|-----------------|-------------------------------------------------------------------------------------------------------|
| `make build`    | Compila el proyecto Go y guarda el binario en la carpeta `bin/`.                                       |
| `make run`      | Ejecuta la aplicación directamente con `go run`.                                                       |
| `make test`     | Ejecuta las pruebas unitarias.                                                                          |
| `make clean`    | Elimina el binario generado en la compilación.                                                          |
| `make lint`     | Ejecuta el linter `golangci-lint` usando la configuración del archivo `.golangci.yml`.                 |

---

### 🧪 **Entorno de Desarrollo (dev)**
| Comando                  | Descripción                                                                                           |
|--------------------------|-------------------------------------------------------------------------------------------------------|
| `make dev-build`         | Construye los servicios en modo desarrollo usando `docker-compose.dev.yml`.                           |
| `make dev-up`            | Inicia los servicios en modo desarrollo.                                                               |
| `make dev-down`          | Detiene los servicios en modo desarrollo y elimina contenedores huérfanos.                             |
| `make dev-logs`          | Muestra los logs en tiempo real de los servicios en modo desarrollo.                                    |

> 💡 **Nota:** El perfil de desarrollo (`euxcel-api`) está definido en `docker-compose.dev.yml`.

---

### 🌐 **Entorno de Staging (stg)**
| Comando                  | Descripción                                                                                           |
|--------------------------|-------------------------------------------------------------------------------------------------------|
| `make stg-build`         | Construye los servicios en modo staging usando `docker-compose.stg.yml`.                              |
| `make stg-up`            | Inicia los servicios en modo staging.                                                                  |
| `make stg-down`          | Detiene los servicios en modo staging y elimina contenedores huérfanos.                                |
| `make stg-logs`          | Muestra los logs en tiempo real de los servicios en modo staging.                                       |

> 💡 **Nota:** El perfil de staging (`euxcel-api`) está definido en `docker-compose.stg.yml`.

---

## 📂 **Estructura del proyecto**
```plaintext
├── bin/                        # Binarios generados después de la compilación
├── cmd/                        # Código fuente del proyecto
├── docker-compose.dev.yml      # Configuración de Docker para entorno de desarrollo
├── docker-compose.stg.yml      # Configuración de Docker para entorno de staging
├── init-mongo.js               # Script de inicialización de MongoDB
├── Makefile                    # Archivo de tareas Makefile
└── .golangci.yml               # Configuración del linter
```

---

## 📝 **Ejemplo de ejecución**

1. **Construir e iniciar en entorno de desarrollo:**
```bash
make dev-build
make dev-up
```

2. **Ver logs en entorno de desarrollo:**
```bash
make dev-logs
```

3. **Detener los servicios en entorno de desarrollo:**
```bash
make dev-down
```

4. **Construir y ejecutar pruebas:**
```bash
make build
make test
```

---

## 🎯 **Consejos**
- Asegúrate de que el script `init-mongo.js` tenga permisos de lectura.
- Si MongoDB no se está iniciando correctamente, verifica los logs con:
```bash
docker logs mongodb
```
- Si hay conflictos de puertos en Docker, intenta detener otros servicios o cambiar el puerto en el archivo `docker-compose`.

---

## 🏆 **¡Listo para construir y ejecutar el proyecto!** 😎