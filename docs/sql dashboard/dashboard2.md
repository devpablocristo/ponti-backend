Avance de siembra
Avance de costos
Avance de cosecha
Avance de aportes
Resultado operativo

Balance de gestión
Incidencia de costos por cultivo
Indicadores operativos

filtros
customer_id <--- todos los campos de 1 cliente
project_id <--- todos los campos de un proyecto
campaign_id <---  todos los campos de una campaña
field_id: Campo <--- un campo en particular


NOTA:

Lista de Datos Hardcodeados:

1. budget_total_usd = $20,000
Presupuesto total por proyecto (hardcodeado)

2. `contributions_progress_pct = 100.00

Motivo:
Propósito: Confirmar que el proyecto tiene participación completa de inversores
Lógica: Los inversores SIEMPRE deben sumar 100% por proyecto
Estado: **Hardcodeado necesario** para validación
Comentario**: SIEMPRE 100% por proyecto

**Solo quedan 2 valores hardcodeados** y ambos son **NECESARIOS**:
1. **$20,000** para presupuesto (funcionalidad)
2. **100.00** para validación (integridad)

###########################################################################################################################
###########################################################################################################################

### 1. **Avance de Siembra** (líneas 95-110)
- **CTE**: `sowing`
- **Campos**: `sowing_hectares`, `sowing_total_hectares`
- **Cálculo**: Área sembrada vs área total por proyecto/campaña/field



Ejemplo de Cálculo:
Si tienes:
Lote 1: 10 hectáreas con fecha de siembra → cuenta como sembrado
Lote 2: 15 hectáreas sin fecha de siembra → cuenta como 0 sembrado
Lote 3: 20 hectáreas con fecha de siembra → cuenta como sembrado



Resultado:
sowed_area = 30 hectáreas (10 + 0 + 20)
total_hectares = 45 hectáreas (10 + 15 + 20)
Porcentaje de avance = (30/45) × 100 = 66.67%



###########################################################################################################################
###########################################################################################################################

### 2. **Avance de Costos** (líneas 129-150)
- **CTE**: `costs_agg`
- **Campos**: 
  - `executed_costs_usd` (costos directos ejecutados)
  - `executed_labors_usd` (mano de obra ejecutada)
  - `executed_supplies_usd` (insumos utilizados)
  - `budget_cost_usd` (costos administrativos presupuestados)



Ejemplo de Cálculo del Avance de Costos:

Si tienes:

Proyecto 1:
Labor A: $1,000 con workorder ejecutada (effective_area > 0) → cuenta como ejecutado
Labor B: $1,500 sin workorder ejecutada → cuenta como 0 ejecutado
Supply A: $500 con workorder_item utilizado (final_dose > 0) → cuenta como utilizado
Supply B: $750 sin workorder_item utilizado → cuenta como 0 utilizado

Proyecto 2:
Labor C: $800 con workorder ejecutada (effective_area > 0) → cuenta como ejecutado
Supply C: $400 con workorder_item utilizado (final_dose > 0) → cuenta como utilizado

Agregado por Campaña:
executed_costs_usd = 2,700 USD (1,500 + 1,200)
budget_total_usd = 20,000 USD (10,000 + 10,000 por proyecto)

Porcentaje de Avance:
costs_progress_pct = (2,700 / 20,000) × 100 = 13.50%

Campos en la Vista:
executed_costs_usd:     2700    -- Costos ejecutados totales
executed_labors_usd:    1800    -- Labores ejecutadas (1000 + 800)
executed_supplies_usd:  900     -- Insumos utilizados (500 + 400)
budget_total_usd:       20000   -- Presupuesto total hardcodeado
costs_progress_pct:     13.50   -- Porcentaje de avance

Interpretación:
Costos Ejecutados: $2,700 (13.50% del presupuesto)
Presupuesto Total: $20,000
Avance de Costos: 13.50%
Estado: Muy por debajo del presupuesto (bien controlado)

###########################################################################################################################
###########################################################################################################################

### 3. **Avance de Cosecha** (líneas 112-127)
- **CTE**: `harvest`
- **Campos**: `harvest_hectares`, `harvest_total_hectares`
- **Cálculo**: Área cosechada vs área total por proyecto/campaña/field


Ejemplo de Cálculo:
Si tienes:
Lote 1: 10 hectáreas con 5 toneladas cosechadas → cuenta como cosechado
Lote 2: 15 hectáreas con 0 toneladas → cuenta como 0 cosechado
Lote 3: 20 hectáreas con 8 toneladas cosechadas → cuenta como cosechado
Lote 4: 12 hectáreas sin toneladas (NULL) → cuenta como 0 cosechado

Resultado:
harvested_area = 30 hectáreas (10 + 0 + 20 + 0)
total_hectares = 57 hectáreas (10 + 15 + 20 + 12)
Porcentaje de avance = (30/57) × 100 = 52.63%


###########################################################################################################################
###########################################################################################################################


### 4. **Avance de Aportes** (línea 235)
- **Campo**: `contributions_progress_pct`


## **Ejemplo de Cálculo del Avance de Aportes:**

### **Si tienes:**

#### **Cliente A (customer_id=100):**
- **Proyecto 1** (project_id=1):
  - **Campaña 200** (campaign_id=200):
    - **Field 10**: 50 hectáreas
    - **Field 11**: 75 hectáreas
  - **Campaña 201** (campaign_id=201):
    - **Field 12**: 60 hectáreas
- **Proyecto 2** (project_id=2):
  - **Campaña 200** (campaign_id=200):
    - **Field 20**: 80 hectáreas

#### **Inversores por Proyecto:**
```
Proyecto 1:
- Inversor 1: 40% participación
- Inversor 2: 35% participación
- Inversor 3: 25% participación

Proyecto 2:
- Inversor 4: 60% participación
- Inversor 5: 40% participación
```

### **Resultado:**

#### **Por Field (Nivel más detallado):**
```
customer_id=100, project_id=1, campaign_id=200, field_id=10:
- investor_id: 1, investor_name: "Inversor 1", investor_percentage_pct: 40.00, contributions_progress_pct: 100.00
- investor_id: 2, investor_name: "Inversor 2", investor_percentage_pct: 35.00, contributions_progress_pct: 100.00
- investor_id: 3, investor_name: "Inversor 3", investor_percentage_pct: 25.00, contributions_progress_pct: 100.00

customer_id=100, project_id=1, campaign_id=200, field_id=11:
- investor_id: 1, investor_name: "Inversor 1", investor_percentage_pct: 40.00, contributions_progress_pct: 100.00
- investor_id: 2, investor_name: "Inversor 2", investor_percentage_pct: 35.00, contributions_progress_pct: 100.00
- investor_id: 3, investor_name: "Inversor 3", investor_percentage_pct: 25.00, contributions_progress_pct: 100.00

customer_id=100, project_id=1, campaign_id=201, field_id=12:
- investor_id: 4, investor_name: "Inversor 4", investor_percentage_pct: 60.00, contributions_progress_pct: 100.00
- investor_id: 5, investor_name: "Inversor 5", investor_percentage_pct: 40.00, contributions_progress_pct: 100.00

customer_id=100, project_id=2, campaign_id=200, field_id=20:
- investor_id: 6, investor_name: "Inversor 6", investor_percentage_pct: 50.00, contributions_progress_pct: 100.00
- investor_id: 7, investor_name: "Inversor 7", investor_percentage_pct: 50.00, contributions_progress_pct: 100.00
```

#### **Por Campaña (Nivel intermedio):**
```
customer_id=100, project_id=1, campaign_id=200:
- investor_id: 1, investor_name: "Inversor 1", investor_percentage_pct: 40.00, contributions_progress_pct: 100.00
- investor_id: 2, investor_name: "Inversor 2", investor_percentage_pct: 35.00, contributions_progress_pct: 100.00
- investor_id: 3, investor_name: "Inversor 3", investor_percentage_pct: 25.00, contributions_progress_pct: 100.00

customer_id=100, project_id=1, campaign_id=201:
- investor_id: 4, investor_name: "Inversor 4", investor_percentage_pct: 60.00, contributions_progress_pct: 100.00
- investor_id: 5, investor_name: "Inversor 5", investor_percentage_pct: 40.00, contributions_progress_pct: 100.00
```

#### **Por Proyecto (Nivel proyecto):**
```
customer_id=100, project_id=1:
- investor_id: 1, investor_name: "Inversor 1", investor_percentage_pct: 40.00, contributions_progress_pct: 100.00
- investor_id: 2, investor_name: "Inversor 2", investor_percentage_pct: 35.00, contributions_progress_pct: 100.00
- investor_id: 3, investor_name: "Inversor 3", investor_percentage_pct: 25.00, contributions_progress_pct: 100.00
- investor_id: 4, investor_name: "Inversor 4", investor_percentage_pct: 60.00, contributions_progress_pct: 100.00
- investor_id: 5, investor_name: "Inversor 5", investor_percentage_pct: 40.00, contributions_progress_pct: 100.00

customer_id=100, project_id=2:
- investor_id: 6, investor_name: "Inversor 6", investor_percentage_pct: 50.00, contributions_progress_pct: 100.00
- investor_id: 7, investor_name: "Inversor 7", investor_percentage_pct: 50.00, contributions_progress_pct: 100.00
```

#### **Por Cliente (Nivel cliente):**
```
customer_id=100:
- investor_id: 1, investor_name: "Inversor 1", investor_percentage_pct: 40.00, contributions_progress_pct: 100.00
- investor_id: 2, investor_name: "Inversor 2", investor_percentage_pct: 35.00, contributions_progress_pct: 100.00
- investor_id: 3, investor_name: "Inversor 3", investor_percentage_pct: 25.00, contributions_progress_pct: 100.00
- investor_id: 4, investor_name: "Inversor 4", investor_percentage_pct: 60.00, contributions_progress_pct: 100.00
- investor_id: 5, investor_name: "Inversor 5", investor_percentage_pct: 40.00, contributions_progress_pct: 100.00
- investor_id: 6, investor_name: "Inversor 6", investor_percentage_pct: 50.00, contributions_progress_pct: 100.00
- investor_id: 7, investor_name: "Inversor 7", investor_percentage_pct: 50.00, contributions_progress_pct: 100.00
```

## **�� Campos en la Vista:**

```sql
-- Para cada nivel:
investor_id:                   1, 2, 3, 4, 5, 6, 7
investor_name:                 "Inversor 1", "Inversor 2", "Inversor 3", "Inversor 4", "Inversor 5", "Inversor 6", "Inversor 7"
investor_percentage_pct:       40.00, 35.00, 25.00, 60.00, 40.00, 50.00, 50.00
contributions_progress_pct:    100.00, 100.00, 100.00, 100.00, 100.00, 100.00, 100.00
```

## **�� Características Clave:**

### **1. Inversores por PROYECTO:**
- **Todos los campos del mismo proyecto** tienen **los mismos inversores**
- **Se repite la información** de inversores para cada campo

### **2. `contributions_progress_pct` siempre 100%:**
- **Confirma** que el proyecto tiene participación completa
- **No es un cálculo**, es una validación

### **3. Estructura Jerárquica:**
- **customer_id**: Todos los inversores del cliente
- **project_id**: Todos los inversores del proyecto
- **campaign_id**: Todos los inversores de la campaña
- **field_id**: Los mismos inversores del proyecto (repetidos)

### **4. Uso en la App:**
- **`investor_percentage_pct`**: Para mostrar el breakdown individual
- **`contributions_progress_pct`**: Para confirmar participación completa
- **Agrupación**: Por nivel de detalle requerido


###########################################################################################################################
###########################################################################################################################


### 5. **Resultado Operativo** (líneas 152-175)
- **CTE**: `operating_result`
- **Campos**:
  - `income_usd` (ingresos por toneladas * $200)
  - `operating_result_usd` (resultado en USD)
  - `operating_result_pct` (porcentaje de rentabilidad)

¡Exacto! Tienes razón. Es mucho más simple de lo que estaba complicando.

## porcentaje rojo grande

**SÍ, se hace todo eso con la fórmula que mencionaste:**

```
Resultado Operativo = Ingresos por comercializaciones - Total invertido
Porcentaje de Rentabilidad = (Resultado Operativo / Total invertido) × 100
```

¡Perfecto! 🎯 Te explico de manera simple y completa cómo se obtienen esos 3 valores:

## 🌱 **¿Cómo Se Calculan los 3 Valores del Dashboard?**

### **1. `total_costs_usd` ($93,500) - Los Costos Totales**

**¿Qué es?** La suma de todo lo que se ha gastado en el proyecto.

**¿De dónde viene?**
- **Labores ejecutadas**: Cada trabajo realizado en el campo (siembra, fertilización, cosecha)
- **Insumos utilizados**: Fertilizantes, semillas, pesticidas que se compraron y usaron
- **Fórmula**: `Labores + Insumos = Costos Totales`

**Ejemplo práctico:**
- Siembra: $2,500 (50 ha × $50/ha)
- Fertilización: $3,750 (50 ha × $75/ha)  
- Cosecha: $5,000 (50 ha × $100/ha)
- **Total**: $11,250 por lote

### **2. `result_usd` (-$45,500) - El Resultado Operativo**

**¿Qué es?** Cuánto dinero ganaste o perdiste realmente.

**¿De dónde viene?**
- **Ingresos**: Dinero que obtienes por las ventas/comercializaciones
- **Costos**: Todo lo que gastaste (labores + insumos)
- **Costo administrativo**: Gastos de oficina, gestión, etc.
- **Fórmula**: `Ingresos - Costos - Costo Administrativo = Resultado`

**Ejemplo práctico:**
- Ingresos del proyecto: $48,000
- Costos totales: $93,500
- Costo administrativo: $0
- **Resultado**: $48,000 - $93,500 = **-$45,500** (pérdida)

### **3. `progress_pct` (-48.7%) - El Porcentaje de Rentabilidad**

**¿Qué es?** Qué tan rentable es tu proyecto en porcentaje.

**¿De dónde viene?**
- **Resultado operativo**: Lo que ganaste o perdiste
- **Total invertido**: Costos + Costo administrativo
- **Fórmula**: `(Resultado / Total Invertido) × 100`

**Ejemplo práctico:**
- Resultado: -$45,500
- Total invertido: $93,500
- **Porcentaje**: (-$45,500 ÷ $93,500) × 100 = **-48.7%**

## �� **Relaciones en la Base de Datos:**

### **Tabla `projects`**
- Contiene el **costo administrativo** de cada proyecto
- Es el proyecto principal que agrupa todo

### **Tabla `workorders` + `workorder_items`**
- Registra cada **labor ejecutada** (siembra, fertilización, cosecha)
- Multiplica hectáreas × precio por hectárea = **costo de labores**

### **Tabla `supplies` + `workorder_items`**
- Registra cada **insumo utilizado** (fertilizantes, semillas)
- Multiplica cantidad × precio = **costo de insumos**

### **Tabla `fields` + `lease_types`**
- Define cómo se calculan los **ingresos** según el tipo de arriendo
- Cada campo tiene un tipo de arriendo que determina cuánto dinero genera

### **Tabla `crop_commercializations`**
- Registra las **ventas/comercializaciones** de cada cultivo
- Es la fuente de los ingresos del proyecto

## 🎯 **Resumen Simple:**

1. **Costos**: Sumas todo lo que gastaste en labores e insumos
2. **Resultado**: Restas ingresos menos costos menos administrativo
3. **Porcentaje**: Divides resultado entre total invertido y multiplicas por 100

**¡Es como hacer las cuentas de tu negocio!** 📊

###########################################################################################################################
###########################################################################################################################

### 6. **Balance de Gestión**
- **Campo**: `operating_result_total_costs_usd` (línea 233)
- **Cálculo**: Costos ejecutados + costos administrativos (B + C)

Costos Directos Ejecutados = Semilla + Insumos + Labores (solo lo ejecutado)
Costos Directos Invertidos = Insumos + Labores (ejecutados o no ejecutados)
Costos Directos Stock = Costos Directos Invertidos - Costos Directos Ejecutados

Desgloce de Costos Directos
Semilla Ejecutados = Todos los insumos ejecutados que sean solo de tipo "Semilla".
Semilla Invertidos = Todos los insumos no ejecutados que sean solo de tipo "Semilla".
Semilla Stock = Semilla Invertidos - Semilla Ejecutados.

Insumos Ejecutados = Todos los insumos ejecutados que no sean de tipo "Semilla".
Insumos Invertidos = Todos los insumos no ejecutados que no sean de tipo "Semilla".
Insumos Stock = Insumos Invertidos - Insumos Ejecutados.

Labores Ejecutados = Todas las labores ejecutadas.
Labores Invertidos = Todas las labores no ejecutadas.
Labores Stock = Labores Invertidos - Labores Ejecutados. 

Arriendo Ejecutados = no se calcula.
Arriendo Invertidos = Total comercializaciones x 30%.
Arriendo Stock = no se calcula.

Estructura Ejecutados = no se calcula.
Estructura Invertidos = Gastos administrativos fijos del proyecto que se cargan directamente en la entidad Projects como admin_cost.
Estructura Stock = no se calcula.

###########################################################################################################################
###########################################################################################################################


### 7. **Incidencia de Costos por Cultivo**
- **Campos**: 
  - `executed_labors_usd` (mano de obra)
  - `executed_supplies_usd` (insumos)
  - `budget_cost_usd` (administrativos)

Incidencia de Costos por Cultivo:
Filas: Cultivos del proyecto/campo
Columnas: Superficie (Has), Incidencia %, Costo por Ha por cultivo

Cálculos Específicos:
1. Superficie (Has):
Fuente: Suma de superficies asociadas al cultivo
Datos: Siembra por lotes / Info cierre final
Tabla: lots.hectares con current_crop_id

2. Incidencia %:
Fórmula: (Hectáreas del cultivo / Hectáreas totales del proyecto) × 100
Ejemplo: Soja 32 ha de 100 ha totales = 32%

3. Costo por Ha por cultivo:
Fuente: Suma de todo lo invertido en órdenes de trabajo sobre el cultivo
Componentes: Labores + Insumos
Cálculo: Costos directos del cultivo / Hectáreas del cultivo

###########################################################################################################################
###########################################################################################################################


### 8. **Indicadores Operativos**
- **Porcentaje de rentabilidad**: `operating_result_pct`
- **Área sembrada vs total**: `sowing_hectares` / `sowing_total_hectares`
- **Área cosechada vs total**: `harvest_hectares` / `harvest_total_hectares`



Primera orden de trabajo
Fecha y número de la primera actividad registrada en campo. Marca inicio oficial de la campaña. Útil para cronogramas y control de tiempos.
Última orden de trabajo
Fecha y número de la última orden cargada. Indica actividad más reciente y si hay labores pendientes o finalizadas.
Arqueo de stock
Fecha en que se cerró el último inventario de insumos, registrando existencia física vs. consumo. Permite controlar desvíos, faltantes y mantener trazabilidad.
Cierre de campaña
Fecha estimada o real de finalización del ciclo agrícola. Define cierre del ciclo productivo. Se puede cargar y editar manualmente. (No tenemos este dato Aun)




###########################################################################################################################
###########################################################################################################################