# **Documentación Completa de “QH / Que Hay?!”**

## **Índice General**

1. [Visión General del Proyecto](#visión-general-del-proyecto)  
2. [Arquitectura y MVP (QH MVP)](#arquitectura-y-mvp-qh-mvp)  
   2.1. [Arquitectura](#arquitectura)  
   2.2. [Microservicios](#microservicios)  
   2.3. [Tecnologías Principales](#tecnologías-principales)  
   2.4. [Puesta en Marcha (Getting Started)](#puesta-en-marcha-getting-started)  
   2.5. [Futuras Mejoras (Future Enhancements)](#futuras-mejoras-future-enhancements)  
   2.6. [Contribuciones (Contributing)](#contribuciones-contributing)  
   2.7. [Licencia (License)](#licencia-license)  

3. [Plan de Proyecto “Que Hay?!”](#plan-de-proyecto-que-hay)  
   3.1. [Resumen Ejecutivo](#resumen-ejecutivo)  
   3.2. [Objetivos del Proyecto](#objetivos-del-proyecto)  
   3.3. [Análisis de Mercado y Público Objetivo](#análisis-de-mercado-y-público-objetivo)  
   3.4. [Características y Funcionalidades](#características-y-funcionalidades)  
   3.5. [Tecnologías Involucradas](#tecnologías-involucradas)  
   3.6. [Desarrollo del MVP](#desarrollo-del-mvp)  
   3.7. [Plan de Desarrollo](#plan-de-desarrollo)  
   3.8. [Seguridad y Privacidad](#seguridad-y-privacidad)  
   3.9. [Estrategias de Marketing y Adquisición de Usuarios](#estrategias-de-marketing-y-adquisición-de-usuarios)  
   3.10. [Roadmap y Futuras Mejoras](#roadmap-y-futuras-mejoras)  

4. [Flujo General de Usuario](#flujo-general-de-usuario)  
   4.1. [Diseño del Flujo de Usuario](#diseño-del-flujo-de-usuario)  
   4.2. [Onboarding, Registro e Inicio de Sesión](#onboarding-registro-e-inicio-de-sesión)  
   4.3. [Perfil de Usuario](#perfil-de-usuario)  
   4.4. [Descubrimiento y Creación de Eventos](#descubrimiento-y-creación-de-eventos)  
   4.5. [Organización de Reuniones Privadas y Gastos](#organización-de-reuniones-privadas-y-gastos)  
   4.6. [Conexión con Otros Usuarios y Grupos](#conexión-con-otros-usuarios-y-grupos)  
   4.7. [Asistente IA](#asistente-ia)  
   4.8. [Pagos y Transacciones](#pagos-y-transacciones)  
   4.9. [Notificaciones y Recordatorios](#notificaciones-y-recordatorios)  
   4.10. [Configuración de Seguridad](#configuración-de-seguridad)  
   4.11. [Cierre de Sesión](#cierre-de-sesión)  

5. [Flujo Enriquecido: “Elegir un Evento”](#flujo-enriquecido-elegir-un-evento)  
   5.1. [Búsqueda y Listado de Eventos](#búsqueda-y-listado-de-eventos)  
   5.2. [Detalles de un Evento](#detalles-de-un-evento)  
   5.3. [Acciones Principales (Comprar Entradas, Grupos, etc.)](#acciones-principales-comprar-entradas-grupos-etc)  
   5.4. [Creación y Gestión de Grupos](#creación-y-gestión-de-grupos)  
   5.5. [Post-Elección del Evento](#post-elección-del-evento)  
   5.6. [Consideraciones Técnicas y de UX](#consideraciones-técnicas-y-de-ux)  

6. [Conclusión General](#conclusión-general)  

---

## 1. **Visión General del Proyecto**

**Nombre Provisional del Producto:** Que Hay?! (QH)

**Descripción Resumida:**  
“Que Hay?!” es una red social orientada a conectar personas (principalmente de 30 años en adelante) y empresas a través de eventos locales y temáticos. La plataforma permite descubrir, organizar y participar en eventos, además de crear grupos, gestionar gastos en eventos privados y fomentar la interacción segura entre los usuarios. Incorpora un asistente de inteligencia artificial (IA) para recomendaciones y validación de perfiles con tecnología blockchain para mayor confianza.

**Objetivo Principal:**  
Brindar una experiencia única y segura de socialización en la vida real, ayudando a los usuarios a encontrar eventos, asistir con otras personas de forma confiable, y facilitar la organización y la diversión.

---

## 2. **Arquitectura y MVP (QH MVP)**

### 2.1. **Arquitectura**

El proyecto sigue una **arquitectura de microservicios**, en la cual cada funcionalidad principal está encapsulada en un servicio independiente. Esto facilita la escalabilidad, el mantenimiento y la posibilidad de desplegar y actualizar servicios por separado. Un diagrama de ejemplo (placeholder) podría incluir:

- **Gateway o API Gateway** para enrutar solicitudes.  
- **Servicios de Negocio** (Autenticación, Autorización, Usuarios, Eventos, Grupos, Calificaciones, Notificaciones).  
- **Bases de Datos** adaptadas a cada servicio (SQL, NoSQL, caché, colas de mensajería).  
- **Interfaz Frontend** (aplicación móvil o web).  

> *Sugerencia:* Puedes reemplazar el diagrama placeholder con uno real a medida que avances en el desarrollo.

### 2.2. **Microservicios**

A continuación, se describen los principales servicios que conforman el MVP:

#### a) Authentication Service (`Authe`)

- **Función:**  
  Maneja registro de usuarios, login, verificación de credenciales, generación y validación de tokens (JWT, OAuth).
- **Características Clave:**  
  - Soporta registro para individuos y empresas.  
  - Autenticación basada en tokens.  
  - Integración con proveedores externos (Google, Facebook, etc.).  
- **Base de Datos:** PostgreSQL o MongoDB.

#### b) Authorization Service (`Autho`)

- **Función:**  
  Maneja roles y permisos, aplicando control de acceso basado en roles (RBAC).
- **Características Clave:**  
  - Definición de roles (admin, user, etc.).  
  - Asignación de permisos a acciones específicas.  
  - Validación de privilegios para recursos.  
- **Base de Datos:** Redis o sistemas de caché similares.

#### c) User Service (`Users`)

- **Función:**  
  Gestiona los perfiles de usuarios (personas y empresas): creación, modificación, eliminación y recuperación de datos.
- **Características Clave:**  
  - Manejo diferenciado de perfiles individuales y de empresa.  
  - Atributos de usuario (datos personales, info de compañía, verificación).  
  - Escalable para futuras extensiones.  
- **Base de Datos:** PostgreSQL o MongoDB.

#### d) Event Service (`Events`)

- **Función:**  
  Permite crear, modificar y visualizar eventos, además de gestionar publicaciones relacionadas.
- **Características Clave:**  
  - Creación de eventos (fecha, lugar, categoría).  
  - Publicación/promoción de eventos por empresas.  
  - Gestión de publicaciones (anuncios, discusiones).  
- **Base de Datos:** PostgreSQL o Event Store.

#### e) Group Service (`Group`)

- **Función:**  
  Facilita la formación de grupos de usuarios para asistir juntos a eventos.
- **Características Clave:**  
  - Creación y manejo de grupos.  
  - Asociación de grupos con eventos.  
  - Roles dentro de los grupos (admin, miembro).  
- **Base de Datos:** PostgreSQL o MongoDB.

#### f) Rating Service (`Rating`)

- **Función:**  
  Ofrece un sistema de calificaciones y comentarios sobre eventos y empresas.
- **Características Clave:**  
  - Calificaciones (estrellas, puntuación) y comentarios.  
  - Agregación de puntuaciones para eventos/empresas.  
  - Moderación de contenido inapropiado.  
- **Base de Datos:** MongoDB.

#### g) Notification Service (`Notification`)

- **Función:**  
  Envía notificaciones a los usuarios sobre eventos, invitaciones de grupo y valoraciones recibidas.
- **Características Clave:**  
  - Notificaciones vía email, push u otros canales.  
  - Alertas sobre actualizaciones y acciones relevantes.  
  - Gestión de preferencias y horarios de notificación.  
- **Base de Datos:** Redis o RabbitMQ (para colas de mensajería).

### 2.3. **Tecnologías Principales**

- **Lenguaje de Programación:** Go (Golang).  
- **Frameworks:** Gin-Gonic o Echo para APIs.  
- **Bases de Datos:** PostgreSQL, MongoDB, Redis, Cassandra (para mensajería).  
- **Mensajería y Colas:** RabbitMQ, Kafka (opcional).  
- **Autenticación:** JWT, OAuth 2.0.  
- **Containerización:** Docker y Docker Compose.  
- **Orquestación (opcional):** Kubernetes.  
- **Monitoreo y Logging:** Prometheus, Grafana, ELK Stack.  
- **Control de Versiones:** Git.  
- **CI/CD:** GitHub Actions, Jenkins (opcional).  
- **Otras Herramientas:** Viper (config), Go-Micro (microservicios).

### 2.4. **Puesta en Marcha (Getting Started)**

#### Prerrequisitos

- **Go** >= 1.22  
- **Docker** y **Docker Compose**  
- **Git**  
- **Node.js** y **npm** (si se incluye frontend)  
- **Sistemas de Base de Datos:** PostgreSQL, MongoDB, Redis  

#### Pasos de Instalación

1. **Clonar el Repositorio:**
   ```bash
   git clone https://github.com/yourusername/event-social-network-mvp.git
   cd event-social-network-mvp
   ```
2. **Configurar Variables de Entorno:**
   - Crear un archivo `.env` con las credenciales y valores necesarios para cada microservicio.
3. **Levantar Microservicios con Docker Compose:**
   ```bash
   docker-compose up -d
   ```
4. **Acceder a los Servicios:**
   - Cada servicio estará disponible en el puerto configurado dentro de `docker-compose.yml`.

#### Desarrollo Local

1. **Ir a la Carpeta del Microservicio:**
   ```bash
   cd services/authentication
   ```
2. **Ejecutar con Go:**
   ```bash
   go run main.go
   ```
3. **Pruebas:**
   ```bash
   go test ./...
   ```

#### Documentación de APIs

- Se recomienda usar Swagger, Postman u OpenAPI para documentar cada microservicio.

### 2.5. **Futuras Mejoras (Future Enhancements)**

- **Separación de Servicios de Usuario:** Dividir servicios para usuarios individuales y empresas.  
- **Chat en Tiempo Real:** Servicio de mensajería para conversar dentro de la plataforma.  
- **Servicio de Pagos:** Para eventos de pago o funciones premium.  
- **Servicio de Analítica:** Para recolectar y analizar datos de uso.  
- **Servicio de Multimedia:** Para subir y almacenar fotos y videos.  
- **Servicio de Búsqueda:** Integrar ElasticSearch para búsquedas avanzadas.

### 2.6. **Contribuciones (Contributing)**

1. **Forkea el Repositorio**  
2. **Crea una Rama de Funcionalidad**  
   ```bash
   git checkout -b feature/nueva-funcionalidad
   ```
3. **Realiza tus Cambios y Haz Commit**  
   ```bash
   git commit -m "Agrega nueva funcionalidad"
   ```
4. **Haz Push a la Rama**  
   ```bash
   git push origin feature/nueva-funcionalidad
   ```
5. **Abre un Pull Request**

### 2.7. **Licencia (License)**

Este proyecto está bajo la [Licencia MIT](LICENSE).

---

## 3. **Plan de Proyecto “Que Hay?!”**

### 3.1. **Resumen Ejecutivo**

**Nombre del Producto (Provisional):** Que Hay?!

**Descripción:**  
Que Hay?! es una red social diseñada para conectar personas —especialmente de 30, 40, 50 años o más— a través de eventos locales, facilitando la organización y participación en diferentes actividades. La aplicación permite crear reuniones privadas, asignar roles y gestionar gastos, integrar pagos mediante Mercado Pago y mejorar la experiencia con un asistente de IA para recomendaciones y seguridad.

**Objetivo Principal:**  
Validar rápidamente la propuesta de valor creando un MVP que permita a los usuarios descubrir y asistir a eventos, así como socializar de manera segura y confiable.

### 3.2. **Objetivos del Proyecto**

1. **Conectar Personas:** Proveer herramientas para que los usuarios formen grupos y asistan juntos a eventos.  
2. **Facilitar la Organización de Eventos:** Brindar funcionalidades para crear y gestionar eventos, dividir gastos y asignar roles.  
3. **Recomendaciones con IA:** Sugerir eventos y contactos compatibles.  
4. **Garantizar Seguridad:** Implementar verificación de usuarios (blockchain, 2FA), moderación de contenido y funcionalidades de emergencia.  
5. **Validar Identidades:** Incrementar la confianza a través de procesos de verificación descentralizados y transparentes.

### 3.3. **Análisis de Mercado y Público Objetivo**

- **Edad Principal:** 30+, 40+, 50+ (sin excluir a otros grupos).  
- **Intereses:** Conciertos, deportes, tecnología, cultura, networking, etc.  
- **Necesidades Clave:**
  - Información relevante de eventos.  
  - Compañía confiable para asistir.  
  - Facilidad de uso y seguridad en las interacciones.  
  - Gestión de gastos y roles de manera transparente.

**Análisis de Competencia:**
- **Facebook / Meetup:** No están 100% centrados en la seguridad ni en la gestión de gastos colaborativa.  
- **Eventbrite / Bandsintown:** Se enfocan más en la venta y promoción de eventos, sin profundizar en la conexión social.

**Ventaja Competitiva de Que Hay?!:**
- Especialización en la interacción social segura y en la gestión colaborativa de eventos.  
- Uso de IA para recomendaciones y seguridad.  
- Verificación de identidades con blockchain.

### 3.4. **Características y Funcionalidades**

1. **Descubrimiento de Eventos:** Listados locales, filtros por categoría, fecha, ubicación y recomendaciones personalizadas.  
2. **Organización de Reuniones Privadas:**  
   - Creación de eventos privados.  
   - Roles para compra de bebidas, comida, etc.  
   - Gestión transparente de gastos, con opción de dividir costos.  
   - Pagos integrados con Mercado Pago.  
3. **Conexión con Desconocidos:**  
   - Creación de grupos basados en intereses.  
   - Chats dentro de la app.  
   - Verificación de perfiles.  
4. **Asistente IA:**  
   - Recomendaciones de eventos según historial e intereses.  
   - Asistencia en transporte, compra de tickets y alertas de seguridad.  
   - Compatibilidad social: sugerencias de personas afines.  
5. **Valoraciones y Reseñas:**  
   - Sistema de calificaciones para eventos y organizadores.  
   - Comentarios y reseñas visibles en el perfil de los organizadores.  

### 3.5. **Tecnologías Involucradas**

- **Backend:** Golang con Gin/Echo.  
- **Frontend:** React Native (móvil multiplataforma).  
- **Bases de Datos:** PostgreSQL (datos estructurados), MongoDB (datos flexibles).  
- **IA:** Python con PyTorch o API de OpenAI para recomendaciones iniciales.  
- **Blockchain:** Ethereum o cadena privada para verificación de identidad.  
- **Infraestructura:** AWS/Google Cloud (Docker, Kubernetes, CI/CD).  
- **Seguridad y Autenticación:** OAuth 2.0, JWT, cifrado HTTPS, 2FA.  

### 3.6. **Desarrollo del MVP**

**Definición del MVP:**  
Implementar las funcionalidades esenciales que permitan a los usuarios:  
1) Registrarse y autenticarse,  
2) Buscar y crear eventos,  
3) Organizar reuniones privadas,  
4) Conectarse con grupos,  
5) Usar un asistente IA básico para recomendaciones,  
6) Realizar pagos compartidos.

**Pasos para Desarrollar el MVP:**

1. **Planificación y Diseño:**  
   - Diagramar el flujo de usuario y crear wireframes (Figma, Sketch).  
2. **Desarrollo del Backend:**  
   - Implementar microservicios de autenticación, usuarios, eventos y pagos.  
   - Integraciones con Mercado Pago y APIs de terceros (tickets, transporte).  
3. **Desarrollo del Frontend:**  
   - Pantallas en React Native para registro, creación de eventos, gestión de gastos y grupos.  
4. **Integración de IA Básica:**  
   - Conectar con API de OpenAI para sugerencias y asistencia limitada.  
5. **Pruebas y Feedback:**  
   - Pruebas unitarias e integración.  
   - Beta testing con un grupo reducido de usuarios.  
6. **Despliegue Inicial:**  
   - Preparar y publicar la app en tiendas móviles (App Store, Google Play).  
   - Monitorear métricas y feedback.

### 3.7. **Plan de Desarrollo**

- **Fase 1: Investigación y Planificación (1-2 meses)**  
  Definición de requisitos, análisis de mercado, diseño UX/UI, stack tecnológico.  
- **Fase 2: Diseño y Prototipado (1 mes)**  
  Wireframes, validación de flujo y UX con potenciales usuarios.  
- **Fase 3: Desarrollo Backend y Frontend (3-4 meses)**  
  Implementación de APIs, bases de datos, integración con pagos y el frontend móvil.  
- **Fase 4: Integración de IA y Funcionalidades Avanzadas (2 meses)**  
  Recomendaciones personalizadas, asistente virtual, algoritmos iniciales.  
- **Fase 5: Pruebas y Beta Testing (1-2 meses)**  
  Depuración de errores, mejoras basadas en feedback de usuarios.  
- **Fase 6: Despliegue y Lanzamiento (1 mes)**  
  Publicación en tiendas, estrategias de marketing inicial.

### 3.8. **Seguridad y Privacidad**

1. **Autenticación y Autorización:**  
   - Uso de OAuth 2.0, JWT, 2FA.  
2. **Cifrado de Datos:**  
   - HTTPS en todas las comunicaciones.  
   - Cifrado de datos sensibles en la BD.  
3. **Verificación de Identidad (Blockchain):**  
   - Proceso descentralizado para validar perfiles.  
4. **Cumplimiento Normativo (GDPR, CCPA, etc.):**  
   - Políticas de privacidad transparentes, posibilidad de eliminación de datos.  
5. **Moderación y Protección:**  
   - Firewalls, detección de intrusos, auditorías de seguridad periódicas.

### 3.9. **Estrategias de Marketing y Adquisición de Usuarios**

1. **Redes Sociales:** Publicidad en Facebook, Instagram, LinkedIn.  
2. **Contenido de Valor (Blogs, Videos):** Explicar beneficios de la app y casos de uso.  
3. **Publicidad Dirigida:** Anuncios segmentados en Google Ads.  
4. **Colaboraciones y Alianzas:** Organizadores de eventos, influencers, asociaciones comunitarias.  
5. **Retención:**  
   - UX de calidad, notificaciones personalizadas, programa de referidos.  
   - Eventos de lanzamiento y promociones en la app.

### 3.10. **Roadmap y Futuras Mejoras**

- **Fase 1: Lanzamiento del MVP**  
  Validar propuesta de valor, medir métricas clave.  
- **Fase 2: Expansión Geográfica y Escalabilidad**  
  Abrir el servicio a más ciudades y países, optimizar infraestructura.  
- **Fase 3: IA Propia**  
  Desarrollar sistemas de aprendizaje con datos internos.  
- **Fase 4: Blockchain para Verificación**  
  Implementar la verificación descentralizada de manera más robusta.  
- **Fase 5: Nuevas Funcionalidades**  
  Integraciones de transporte, venta de tickets, multimedia, búsqueda avanzada.

---

## 4. **Flujo General de Usuario**

### 4.1. **Diseño del Flujo de Usuario**

El flujo de usuario está centrado en la facilidad de descubrimiento de eventos y la interacción social confiable. A continuación, se presenta una descripción de las fases clave:

1. **Onboarding y Registro**  
2. **Inicio de Sesión**  
3. **Perfil de Usuario**  
4. **Descubrimiento de Eventos**  
5. **Detalles de Evento y Unirse**  
6. **Organización de Reuniones Privadas**  
7. **Gestión de Roles y Gastos**  
8. **Conexión con Otros Usuarios**  
9. **Asistente IA**  
10. **Pagos y Transacciones**  
11. **Configuración de Seguridad**  
12. **Notificaciones y Recordatorios**  
13. **Cierre de Sesión**

### 4.2. **Onboarding, Registro e Inicio de Sesión**

1. **Pantalla de Bienvenida:** Información breve de la app y botón de “Comenzar”.  
2. **Opciones de Registro:** Email, Google, Facebook, Apple ID.  
3. **Verificación de Email (si aplica):** Confirmar cuenta a través de un enlace.  
4. **Configuración Inicial del Perfil:** Foto, intereses, rango de edad, ubicación.  
5. **Tutorial Opcional:** Explica funcionalidades básicas.  
6. **Inicio de Sesión:** Campos de email/contraseña o redes sociales. (Opción de 2FA si está activada).

### 4.3. **Perfil de Usuario**

- **Información Personal:** Nombre, foto, edad, ubicación, intereses.  
- **Preferencias y Privacidad:** Configurar notificaciones, categorías de eventos favoritos, visibilidad de perfil.  
- **Verificación de Identidad:** Posibilidad de cargar documentos o usar blockchain.  
- **Historial de Eventos y Reseñas:** Consultar eventos asistidos o calificados.

### 4.4. **Descubrimiento y Creación de Eventos**

- **Pantalla Principal (Dashboard):** Listado de eventos destacados y recomendados.  
- **Búsqueda con Filtros:** Categoría, ubicación, fecha, popularidad, seguridad.  
- **Creación de Eventos Privados o Públicos:** Definir detalles (fecha, lugar, cupo, rol organizador).

### 4.5. **Organización de Reuniones Privadas y Gastos**

- **Creación de Evento Privado:** Selección de participantes, roles (bebidas, comida, etc.), notas de organización.  
- **Gestión de Gastos:** Registro de costos, asignación de quién pagó qué, división automática.  
- **Pagos Integrados:** Mercado Pago u otros para saldar deudas internas.

### 4.6. **Conexión con Otros Usuarios y Grupos**

- **Búsqueda de Grupos:** Filtros por intereses, edad, ubicación.  
- **Chat de Grupo:** Comunicación previa al evento.  
- **Verificación de Perfiles:** Mostrar usuarios con sellos de confianza (blockchain, 2FA, etc.).

### 4.7. **Asistente IA**

- **Recomendaciones de Eventos:** Basadas en historial e intereses.  
- **Asistencia en Reservas y Transporte:** Sugerencias de Uber, tickets, etc.  
- **Monitoreo de Seguridad:** Alertas de riesgo, soporte en tiempo real.  
- **Compatibilidad Social:** Presentar usuarios con afinidades para asistir juntos.

### 4.8. **Pagos y Transacciones**

- **Métodos de Pago:** Mercado Pago, tarjetas, billeteras digitales.  
- **Registro de Gastos:** Resumen de quién pagó y cuánto se debe.  
- **Historial de Transacciones:** Lista de todas las transacciones realizadas.

### 4.9. **Notificaciones y Recordatorios**

- **Tipos de Notificaciones:** Actualizaciones de eventos, invitaciones a grupos, calificaciones, recordatorios.  
- **Configuración:** Ajuste de preferencia y método de notificación (push, email, SMS).  
- **Recordatorios de Eventos:** Alertas unas horas o días antes del evento.

### 4.10. **Configuración de Seguridad**

- **Autenticación de Dos Factores (2FA):** Activa/Desactiva.  
- **Gestión de Dispositivos Conectados:** Monitorear sesiones activas.  
- **Verificación de Identidad con Blockchain:** Proceso para usuarios que buscan mayor nivel de confianza.

### 4.11. **Cierre de Sesión**

- **Pantalla de Configuración:** Opción “Cerrar Sesión”.  
- **Confirmación:** Mensaje de confirmación y fin de la sesión actual.

---

## 5. **Flujo Enriquecido: “Elegir un Evento”**

A continuación se detalla el proceso específico que sigue un usuario cuando decide **elegir y participar** en un evento, abarcando todas las interacciones posibles.

### 5.1. **Búsqueda y Listado de Eventos**

1. **Pantalla de Búsqueda:** Barra para ingresar palabras clave, filtros avanzados (categoría, fecha, ubicación).  
2. **Aplicación de Filtros:** Ordenar por popularidad, cercanía, fecha.  
3. **Listado de Resultados:** Tarjetas de eventos con info básica (nombre, fecha, imagen, ubicación).

### 5.2. **Detalles de un Evento**

- **Información Completa:** Nombre, descripción, fecha/hora, mapa con ubicación, categoría, medidas de seguridad, organizador (empresa o usuario).  
- **Acciones Disponibles:**  
  - Comprar entrada.  
  - Ver o dejar calificaciones y reseñas.  
  - Unirse a un grupo existente o crear uno nuevo.  
  - Compartir evento.  
  - Marcar como favorito.

### 5.3. **Acciones Principales (Comprar Entradas, Grupos, etc.)**

1. **Comprar Entrada:**  
   - Selección de tipo de ticket (general, VIP).  
   - Cantidad, precio total, impuestos.  
   - Pago integrado (Mercado Pago, tarjeta).  
   - Recepción de tickets digitales, confirmación y almacenamiento en la sección “Mis Eventos”.  

2. **Calificaciones y Reseñas:**  
   - Ver reseñas de otros usuarios.  
   - Dejar puntuación (estrellas) y comentario.  

3. **Ver Grupos Existentes:**  
   - Lista de grupos que planean asistir.  
   - Unirse a un grupo (público o privado con aprobación).  

4. **Crear un Grupo Desde Cero:**  
   - Nombre y descripción del grupo.  
   - Invitar amigos o usuarios de la app.  
   - Asignar roles (administrador, moderador).  
   - Configurar privacidad (público/privado).  

5. **Compartir Evento:**  
   - Redes sociales, SMS, email.  
   - Enlace único para invitar amigos.  

6. **Marcar como Favorito:**  
   - Guardar el evento para futura referencia en “Favoritos” dentro del perfil.

### 5.4. **Creación y Gestión de Grupos**

- **Formulario de Grupo:** Nombre, descripción, intereses, límite de miembros.  
- **Invitación de Participantes:** Desde la app, por correo o SMS a contactos externos.  
- **Chat del Grupo:** Coordinación sobre horarios, punto de encuentro, etc.  
- **Roles y Gastos (si es un evento privado):** Asignación de quién se encarga de qué, registro de compras y pagos.

### 5.5. **Post-Elección del Evento**

1. **Confirmación y Preparación:**  
   - Calendario personal actualizado (sincronización con Google/Apple Calendar).  
   - Notificaciones previas al evento.  
2. **Durante el Evento:**  
   - Información en tiempo real (cambios de horario, actualizaciones).  
   - IA puede recomendar rutas, transportes, actividades relacionadas.  
3. **Después del Evento:**  
   - Posibilidad de dejar reseñas.  
   - Subir fotos (futuras funcionalidades de multimedia).  
   - Historial de asistencia.

### 5.6. **Consideraciones Técnicas y de UX**

- **Navegación Intuitiva:** Menú claro, botones con textos descriptivos, secciones bien delimitadas.  
- **Optimización Móvil:** Fluidez, tiempos de carga mínimos, adaptabilidad a pantallas.  
- **Seguridad:** Uso de HTTPS, 2FA opcional, blockchain para verificación de identidades.  
- **Integraciones Externas:** APIs de pagos (Mercado Pago), transporte (Uber), venta de tickets (Ticketmaster/Eventbrite).  
- **Gamificación (Opcional):** Badges por participar en varios eventos, rankings, recompensas.

---

## 6. **Conclusión General**

**Que Hay?!** propone una plataforma integral para conectar personas y empresas alrededor de eventos, impulsada por microservicios escalables, recomendaciones con IA y verificación de identidades mediante blockchain. El enfoque principal recae en la **seguridad**, la **usabilidad** y la **colaboración**, ofreciendo al público objetivo una experiencia confiable y satisfactoria.

A través del **MVP** se validarán las funcionalidades esenciales:  
- Registro y autenticación segura.  
- Descubrimiento y creación de eventos.  
- Organización de reuniones con gastos compartidos.  
- Conexión entre usuarios y creación de grupos.  
- Asistente IA para recomendaciones y seguridad.  
- Gestión de pagos integrada.

El **roadmap** contempla la expansión hacia funcionalidades avanzadas (chat en tiempo real, analítica, verificación total con blockchain, multimedia, búsqueda con ElasticSearch, etc.) y la escalabilidad a nivel global. Con un plan sólido, integración tecnológica robusta y una experiencia de usuario bien diseñada, **Que Hay?!** se posiciona como una solución prometedora para quienes buscan socializar y divertirse a través de eventos reales de manera segura y organizada.