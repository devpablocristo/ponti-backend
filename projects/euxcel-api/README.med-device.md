## **Servicios REST (Comunicación Síncrona)**

### **Microservicios en Golang**

1. **Auth Service**  
   **Función:**  
   - Gestionar la autenticación, registro, generación y validación de tokens (por ejemplo, JWT).  
   - Proveer endpoints como:
     - `POST /auth/login` para iniciar sesión.
     - `POST /auth/register` para crear nuevas cuentas (si se permite el auto-registro).
   **Consideraciones para MVP:**  
   - Implementa validación básica y generación de tokens.
   - Usa middleware para proteger rutas que requieran autenticación.

2. **Patient Service**  
   **Función:**  
   - Gestionar la información clínica y de datos médicos de los pacientes.  
   - Endpoints recomendados:
     - `POST /patients`: Crear un nuevo paciente.
     - `GET /patients`: Listar pacientes.
     - `GET /patients/{id}`: Obtener el detalle de un paciente.
     - `PUT /patients/{id}` / `PATCH /patients/{id}`: Actualizar datos.
     - `DELETE /patients/{id}`: Eliminar un registro.
   **Consideraciones para MVP:**  
   - Utiliza un ORM como GORM para interactuar con la base de datos.
   - Define una separación clara de dominios mediante DTOs y modelos de dominio.

3. **User Service**  
   **Función:**  
   - Gestionar los perfiles de usuario, roles y datos de acceso que pueden ser distintos de los datos clínicos.  
   - Endpoints similares a los de Auth, pero enfocados en la gestión del usuario (por ejemplo, actualizar perfil, cambiar contraseña).
   **Consideraciones para MVP:**  
   - Puede estar estrechamente integrado con el Auth Service o bien existir como servicio separado para mantener la separación de lógica de negocio y autenticación.

4. **Person Service**  
   **Función:**  
   - Manejar información general de la persona (nombre, apellido, contacto, etc.), que puede ser compartida entre Patient y User.  
   - Endpoints para crear, actualizar y consultar la entidad Person.
   **Consideraciones para MVP:**  
   - Permite reutilizar datos y evitar duplicación; se relaciona con los servicios de Patient y User (por ejemplo, vía clave foránea).

5. **Device Service**  
   **Función:**  
   - Registrar y administrar dispositivos IoT que se utilizan para recolectar datos biométricos.  
   - Endpoints:
     - `POST /devices`: Registrar un dispositivo.
     - `GET /devices`: Listar dispositivos.
     - `GET /devices/{id}`: Consultar detalles.
     - `PUT /devices/{id}`: Actualizar configuración.
   **Consideraciones para MVP:**  
   - Asegurar una estructura que permita asociar cada dispositivo a un paciente o grupo de pacientes.

6. **Measurement Service**  
   **Función:**  
   - Recibir datos de mediciones biométricas (por ejemplo, niveles de glucosa, presión arterial).
   - Endpoints:
     - `POST /measurements`: Recibir y almacenar una medición.
     - `GET /measurements`: Listar mediciones.
     - `GET /patients/{id}/measurements`: Listar mediciones asociadas a un paciente.
   **Consideraciones para MVP:**  
   - Después de guardar la medición, publicar un evento asíncrono (a través de RabbitMQ) para otras operaciones (alertas y procesamiento).
   - Validar formatos y rangos de los valores.

---

### **API Gateway con Kong**

1. **API Gateway (Kong)**  
   **Función:**  
   - Actuar como el punto de entrada único para la aplicación, enrutar las solicitudes al microservicio correspondiente y aplicar políticas de seguridad (autenticación, rate-limiting, etc.).  
   **Consideraciones para MVP:**  
   - Utiliza la API de administración de Kong (REST) para configurar rutas y plugins.
   - Permite desacoplar la lógica de routing del código de los microservicios.

---

### **Microservicios en Python**

1. **Reporting Service**  
   **Función:**  
   - Exponer endpoints para generar reportes a demanda y dashboards que muestren estadísticas de datos biométricos (por ejemplo, promedios, tendencias, gráficos).  
   - Endpoints:
     - `GET /reports/patients/{id}` para obtener un reporte detallado.
     - `GET /reports/global` para reportes agregados.
   **Consideraciones para MVP:**  
   - Utiliza frameworks como Flask o FastAPI.
   - Integra librerías para análisis de datos (pandas, matplotlib) si se requiere generar gráficos o informes en tiempo real.
   - Puede leer directamente de la base de datos o de una capa de datos previamente procesados por workers asíncronos.

---

## **Servicios de Comunicación Asíncrona (RabbitMQ)**

### **Microservicios en Golang (Asíncronos)**

1. **Alert/Notification Service**  
   **Función:**  
   - Consumir eventos de mediciones publicadas en RabbitMQ para analizar si se generan alertas (por ejemplo, niveles críticos de glucosa).
   - Enviar notificaciones a los pacientes o al equipo médico.
   **Consideraciones para MVP:**  
   - Implementa un consumidor sencillo que escuche en un tópico o cola designada para "measurement events".
   - Puede utilizar una biblioteca como [streadway/amqp](https://github.com/streadway/amqp) para interactuar con RabbitMQ.

2. **Measurement Background Processor**  
   **Función:**  
   - Procesar datos de mediciones de forma asíncrona, realizando operaciones adicionales (como validaciones, cálculos o enriquecimientos) que no deben bloquear la respuesta al cliente.
   **Consideraciones para MVP:**  
   - Ejecuta tareas de preprocesamiento o reintentos en caso de fallos.
   - Se integra con la cola de RabbitMQ para recibir mensajes y ejecuta tareas en background.

---

### **Microservicios en Python (Asíncronos)**

1. **Analytics/Reporting Worker**  
   **Función:**  
   - Consumir eventos desde RabbitMQ para acumular datos y realizar análisis en batch, generar estadísticas y preparar información para reportes.  
   **Consideraciones para MVP:**  
   - Utiliza frameworks como Celery para gestionar tareas asíncronas.
   - Aprovecha librerías como pandas o scikit-learn para procesar datos.
   - Almacena los resultados en una base de datos optimizada para consultas analíticas (o en archivos, según la complejidad).

2. **Data Processing Worker**  
   **Función:**  
   - Realizar transformaciones adicionales o integraciones entre datos recibidos y otros sistemas (por ejemplo, normalización de series temporales, detección de anomalías).
   **Consideraciones para MVP:**  
   - Puede ser parte del mismo proceso del Analytics Worker o un proceso separado.
   - Su implementación en Python permite iterar rápidamente sobre algoritmos y ajustar modelos.

---

## **Resumen Final para el MVP**

- **Servicios Síncronos en Golang (REST):**  
  - **Auth, Patient, User, Person, Device, Measurement**  
  – Estos servicios se centran en la exposición de endpoints para la interacción directa (CRUD, autenticación, registro de dispositivos y recepción de mediciones).

- **API Gateway (Kong):**  
  - Encaminador de todas las solicitudes, gestión de seguridad y políticas globales a través de una solución externa robusta.

- **Servicios Síncronos en Python:**  
  - **Reporting**  
  – Permite generar reportes y dashboards on-demand utilizando el ecosistema de análisis de datos de Python.

- **Servicios Asíncronos con RabbitMQ:**  
  - **Golang:**  
    - **Alert/Notification:** Procesa eventos para generar alertas y notificaciones.
    - **Measurement Background Processor:** Ejecuta tareas adicionales en segundo plano relacionadas con mediciones.
  - **Python:**  
    - **Analytics/Reporting Worker:** Realiza análisis en batch y genera estadísticas.
    - **Data Processing Worker:** Procesa y transforma datos para análisis avanzados.


Rest:
Golang:
Auth
Patient
User
Person
Device 
Measurement

Kong:
API Gateway

Python:
Reporting

RabbitMQ:

Golang:
Alert/Notification
Measurement Background Processor

Python:
Analytics/Reporting Worker
Data Processing Worker 