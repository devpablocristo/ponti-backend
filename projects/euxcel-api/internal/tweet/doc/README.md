## 1. Comunicación con Servicios Externos

### 1.1. Implementar Circuit Breaker y Mecanismos de Reintentos
- **Objetivo:**  
  - Prevenir la cascada de errores y proteger la aplicación cuando un servicio externo (Cassandra, Redis, RabbitMQ o microservicio de usuarios) se vuelve inestable.
- **Acciones:**  
  - Integrar un **Circuit Breaker** utilizando librerías como [goresilience](https://github.com/slok/goresilience) o implementar patrones propios.
  - Configurar **mecanismos de reintentos** con tiempos de espera progresivos (exponential backoff) para las llamadas a servicios externos.

### 1.2. Externalización de Configuraciones
- **Objetivo:**  
  - Centralizar y facilitar la modificación de parámetros críticos sin necesidad de recompilar la aplicación.
- **Acciones:**  
  - Utilizar variables de entorno o archivos de configuración para parámetros como timeouts, número de workers, URLs y puertos de servicios externos.

---

## 2. Observabilidad y Telemetría

### 2.1. Monitoreo de Métricas
- **Objetivo:**  
  - Recolectar métricas (latencia, tasa de errores, throughput) para cada componente y reaccionar ante comportamientos anómalos.
- **Acciones:**  
  - Integrar **Prometheus** para recolectar y visualizar métricas, y configurar alertas adecuadas.

### 2.2. Tracing Distribuido
- **Objetivo:**  
  - Rastrear el flujo de las peticiones a través de todos los servicios para identificar cuellos de botella y fallos.
- **Acciones:**  
  - Implementar tracing con herramientas como **Jaeger** o **Zipkin**.

### 2.3. Logging Centralizado
- **Objetivo:**  
  - Facilitar el análisis y la depuración centralizando los logs de todos los servicios.
- **Acciones:**  
  - Adoptar un stack de logging centralizado, por ejemplo, **ELK (Elasticsearch, Logstash, Kibana)** o **Loki**.
  - Migrar de `log.Printf` a una solución de logging estructurado como **Zap** o **Logrus**.

---

## 3. Gestión de Acceso y Tráfico

### 3.1. Utilizar un API Gateway
- **Objetivo:**  
  - Centralizar la entrada a la aplicación, aplicando políticas de autenticación, autorización, rate limiting y logging de forma uniforme.
- **Acciones:**  
  - Evaluar e integrar herramientas como **Kong**, **Traefik** o **NGINX**.
  - Facilitar el versionado y enrutamiento de la API, desacoplando la lógica de negocio de los controles de seguridad.

---

## 4. Optimización del Cache y Estrategias de Lectura

### 4.1. Estrategia de Caché (Cache Aside Pattern)
- **Objetivo:**  
  - Mantener el caché actualizado y reducir la carga en la base de datos.
- **Acciones:**  
  - Implementar políticas de expiración y actualización asíncrona en **Redis**.
  - Asegurar que los datos críticos (por ejemplo, timelines) se refresquen sin saturar la base de datos.

### 4.2. Optimización del Fan-out
- **Objetivo:**  
  - Distribuir eficientemente los tweets a los timelines de los seguidores, especialmente en picos de tráfico.
- **Acciones:**  
  - Evaluar el uso de sistemas de colas o servicios de stream processing (por ejemplo, **Kafka** o ampliar el uso de **RabbitMQ**) para manejar grandes volúmenes de operaciones de fan-out.

---

## 5. Diseño Modular y Escalabilidad Horizontal

### 5.1. Arquitectura Basada en Clean Architecture/DDD
- **Objetivo:**  
  - Garantizar una separación clara entre la API, la lógica de negocio, la persistencia, el caché y la mensajería.
- **Acciones:**  
  - Mantener y documentar la separación de responsabilidades para facilitar el mantenimiento y la evolución de cada componente de forma independiente.

### 5.2. Contenerización y Orquestación
- **Objetivo:**  
  - Permitir el escalado horizontal de cada servicio conforme aumente la carga.
- **Acciones:**  
  - Planificar y documentar el despliegue de componentes en **contenedores (Docker)** y su orquestación mediante **Kubernetes**.

---

## 6. Pruebas y Validación

### 6.1. Pruebas Unitarias y de Integración
- **Objetivo:**  
  - Validar la lógica de negocio, la conversión de modelos y el manejo de errores.
- **Acciones:**  
  - Implementar pruebas unitarias para los casos críticos.
  - Desarrollar pruebas de integración para asegurar la correcta interacción con servicios externos (Cassandra, Redis, RabbitMQ).

### 6.2. Testing de Rendimiento
- **Objetivo:**  
  - Simular escenarios de alta carga y validar la capacidad de la solución para soportar millones de usuarios.
- **Acciones:**  
  - Utilizar herramientas como **JMeter** o **k6** para pruebas de carga y rendimiento.

---

## 7. Documentación y Razonamiento Arquitectónico

### 7.1. Documentación Técnica y Diagramas de Arquitectura
- **Objetivo:**  
  - Proveer una visión clara de la arquitectura y las decisiones técnicas para facilitar la comunicación con equipos de desarrollo y operaciones.
- **Acciones:**  
  - Incluir diagramas que ilustren la interacción entre el API Gateway, los servicios de negocio, la base de datos, el caché y el broker.
  - Crear un documento (por ejemplo, `business.txt` o `README.md`) que detalle las decisiones de diseño y las estrategias para soportar cargas masivas.

### 7.2. Consideraciones de Seguridad
- **Objetivo:**  
  - Asegurar que, aunque se asuma la validez de los usuarios en el ejercicio, se tengan en cuenta futuras necesidades de validación y autenticación.
- **Acciones:**  
  - Documentar las asunciones actuales y planificar la incorporación de controles de seguridad adicionales en una versión de producción.

---

## 8. Código y Buenas Prácticas

### 8.1. Manejo de Errores y Logging
- **Objetivo:**  
  - Mejorar la trazabilidad y depuración de la aplicación.
- **Acciones:**  
  - Consolidar el manejo de errores utilizando `fmt.Errorf` y migrar a soluciones de logging estructurado.

### 8.2. Uso del Contexto y Conversión entre Modelos
- **Objetivo:**  
  - Mantener la consistencia y desacoplamiento en la aplicación.
- **Acciones:**  
  - Seguir propagando el `context.Context` en todas las llamadas.
  - Continuar utilizando funciones de conversión (`FromDomain`, `ToDomain`, etc.) para separar el dominio de la infraestructura.

---

