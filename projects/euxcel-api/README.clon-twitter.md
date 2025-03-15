## 1. Requerimientos del Desafío

- **Tweets:**  
  Los usuarios pueden publicar mensajes cortos (máximo 280 caracteres).

- **Follow:**  
  Los usuarios pueden seguir a otros usuarios.  
  **Importante:** Se sigue a un **usuario**, no a un tweet, para que el timeline muestre todas las publicaciones futuras de ese usuario.

- **Timeline:**  
  El timeline es la línea de tiempo que muestra, de forma cronológica (generalmente de los más recientes a los más antiguos), los tweets de los usuarios que un usuario sigue.

- **Suposiciones y Objetivos:**
  - No es necesario implementar autenticación o manejo de sesiones; el identificador del usuario puede enviarse por header, parámetro, etc.
  - La solución debe estar pensada para escalar a millones de usuarios, optimizando especialmente las lecturas.

---

## 2. Arquitectura y Diseño del Sistema

### A. **Enfoque de Arquitectura**

- **Arquitectura Hexagonal / Clean Architecture / Port and Adapters:**  
  Se separa la lógica de negocio (dominio y casos de uso) de las implementaciones de infraestructura (repositorios, servicios de cache, gateways HTTP, etc.). Esto facilita:
  - **Testabilidad:** Probar el core sin depender de servicios externos.
  - **Flexibilidad y Escalabilidad:** Cambiar componentes (p.ej., base de datos o mecanismos de cache) sin afectar la lógica central.

### B. **Capas Principales**

1. **Dominio:**  
   - **Entidades:**  
     - **Tweet:** Contiene ID, UserID, contenido, timestamps, etc.
     - **Follow:** Representa la relación entre el seguidor y el seguido.
     - **User:** Contiene información básica del usuario.
   
2. **Casos de Uso / Servicios de Aplicación:**  
   - **TweetService:** Para crear, actualizar, eliminar y consultar tweets.
   - **FollowService:** Para gestionar las relaciones de seguimiento.
   - **TimelineService:** Para construir el timeline, que incluye:
     - Obtener la lista de usuarios que el usuario sigue.
     - Consultar los tweets de esos usuarios.
     - Ordenar cronológicamente (de más reciente a más antiguo).
     - (Opcional) Cachear el resultado para mejorar la velocidad en lecturas.

3. **Infraestructura / Adaptadores:**  
   - **Repositorios:**  
     - Implementaciones in-memory (para pruebas o prototipos).
     - Adaptadores para bases de datos relacionales (PostgreSQL), NoSQL (Cassandra, DynamoDB) o mecanismos de cache (Redis).
   - **Servicios Adicionales:**  
     - **Cache (Redis):** Para almacenar timelines y reducir la latencia.
     - **Sistemas de Mensajería (Kafka, RabbitMQ):** Para actualizar timelines de forma asíncrona en un modelo push.
     - **Elasticsearch:** Para búsquedas de texto completo y análisis de contenido.

---

## 3. Implementación de Endpoints y Rutas con Gin

- **Estructura de Rutas:**  
  Se separan las rutas en grupos según el nivel de acceso:
  - **Públicas:** Operaciones de solo lectura, como `GET /:id` para obtener un tweet o `GET /timeline` para ver la línea de tiempo.
  - **Validadas / Protegidas:** Operaciones críticas como crear (`POST`), actualizar (`PUT`) y eliminar (`DELETE`) tweets se colocan en rutas protegidas, aplicando middlewares de autenticación/validación.

- **Ejemplo de Rutas:**  
  ```go
  apiVersion := h.gsv.GetApiVersion()
  apiBase := "/api/" + apiVersion + "/tweets"
  publicPrefix := apiBase + "/public"
  validatedPrefix := apiBase + "/validated"
  protectedPrefix := apiBase + "/protected"

  // Rutas públicas
  public := router.Group(publicPrefix)
  {
      public.POST("", h.CreateTweet)         // Crea un tweet (aunque se recomienda mover operaciones críticas a rutas protegidas)
      public.GET("/:id", h.GetTweet)           // Obtiene un tweet por ID
      public.PUT("/:id", h.UpdateTweet)        // Actualiza un tweet (operación crítica, idealmente protegida)
      public.DELETE("/:id", h.DeleteTweet)     // Elimina un tweet (también debería estar protegido)
      public.GET("/timeline", h.GetTimeLine)    // Obtiene el timeline del usuario (puede ser público o validado)
  }
  ```

- **Observación:**  
  Es importante que las rutas que permiten modificar datos (crear, actualizar, eliminar) estén protegidas para garantizar la seguridad. Las rutas de lectura pueden ser públicas o protegidas según la necesidad de personalización (por ejemplo, el timeline del usuario autenticado).

---

## 4. Construcción del Timeline

### A. **Flujo para Construir el Timeline**

1. **Handler:**  
   - Recibe la solicitud `GET /timeline`.
   - Extrae el identificador del usuario (por ejemplo, del header `X-User-ID`).
   - Llama al caso de uso `GetTimeLine`.

2. **Caso de Uso (UseCase):**
   - Llama al repositorio de follows para obtener la lista de IDs de usuarios a los que el usuario sigue.
   - Con esos IDs, consulta el repositorio de tweets para obtener los tweets recientes.
   - Ordena los tweets en orden descendente por fecha.
   - (Opcional) Verifica si existe una versión cacheada en Redis; si no, almacena el timeline cacheado para futuras consultas.

3. **Repositorio y Servicios:**
   - **FollowRepository:** Para recuperar la lista de usuarios seguidos.
   - **TweetRepository:** Para obtener tweets basados en la lista de IDs.
   - **CacheService (Redis):** Para cachear el timeline y reducir la latencia en lecturas.

### B. **Servicios y Tecnologías Recomendadas para Escalabilidad**

- **Bases de Datos Relacionales (PostgreSQL):**  
  Para datos transaccionales críticos (usuarios, follows, tweets con integridad ACID).

- **Bases de Datos NoSQL (Cassandra o DynamoDB):**  
  Para almacenar grandes volúmenes de tweets y timelines, optimizando la escritura y la escalabilidad horizontal.

- **Redis:**  
  Para cachear timelines y otros datos de lectura frecuente, ofreciendo tiempos de respuesta muy bajos.

- **Sistemas de Mensajería (Kafka, RabbitMQ):**  
  Para distribuir eventos (como la publicación de un tweet) y actualizar los timelines de forma asíncrona (modelo push).

- **Elasticsearch:**  
  Para indexar y permitir búsquedas rápidas en el contenido de los tweets.

- **Docker y Kubernetes:**  
  Para la contenedorización y orquestación de la aplicación, permitiendo escalabilidad horizontal y alta disponibilidad.

---

## 5. Consideración Final sobre el Timeline y su Orden

- **Orden de Presentación:**  
  El timeline muestra los tweets más recientes primero, lo cual se asemeja a una estructura LIFO en términos visuales, pero conceptualmente se trata de una lista ordenada cronológicamente. No se usa una pila LIFO tradicional, ya que no se "extraen" elementos, sino que se muestran todos para consulta.

---

## Conclusión

Se ha analizado la arquitectura, diseño y flujo para construir una plataforma simplificada de microblogging con las siguientes funcionalidades:
- Publicar tweets.
- Seguir a otros usuarios.
- Generar un timeline que muestre los tweets de los usuarios seguidos.

La solución debe estar estructurada en capas (dominio, casos de uso, infraestructura) y emplear tecnologías y servicios (como Redis, Cassandra, Kafka, PostgreSQL, Elasticsearch, Docker y Kubernetes) para garantizar escalabilidad y optimización en lecturas, pensando en soportar millones de usuarios.

Este resumen integra tanto la parte teórica como las decisiones prácticas y recomendaciones para implementar un sistema robusto y escalable.