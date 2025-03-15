**Flujo Assessment (Preparación y Envío)**

1. **Inicio de Sesión y Autenticación:**  
   - Frontend → Login.  
   - Authe (pep) → Validar credenciales.  
   - Authe → Generar y entregar token al frontend.

2. **Creación de Entidades y Assessment:**  
   - Frontend → Solicita creación de assessment.  
   - Person → Crear registro de la persona.  
   - Candidate → Crear registro del candidato.  
   - Assessments → Crear assessment y generar link único.

3. **Notificación al Candidato:**  
   - Notification → Enviar email con el link único del assessment.


**Flujo Candidato (Ejecución y Evaluación)**

1. **Acceso al Assessment:**  
   - Email del candidato → Clic en el link único recibido.  
   - Authe → Validar token del assessment.

2. **Generación y Presentación de Preguntas:**  
   - GenerateAssessment (IA, python/langchain) → Generar preguntas y estructurar el assessment.  
   - Frontend → Mostrar preguntas al candidato.

3. **Interacción y Registro de Respuestas:**  
   - Candidato → Responde las preguntas a través del frontend.  
   - Frontend → Enviar respuestas junto con eventos del navegador.

4. **Captura y Monitoreo de Eventos:**  
   - BrowserEvent → Captura y guarda eventos del navegador para análisis de comportamiento y seguridad.

5. **Validación, Evaluación y Finalización:**  
   - Assessment → Validar que el assessment esté completo y sin inconsistencias.  
   - Assessment → Evaluar (scoring) las respuestas y asignar puntaje.  
   - Assessment → Marcar assessment como realizado.

6. **Feedback y Notificación Final:**  
   - Notification → Enviar al candidato (y/o al evaluador) confirmación de finalización, puntaje y feedback.
