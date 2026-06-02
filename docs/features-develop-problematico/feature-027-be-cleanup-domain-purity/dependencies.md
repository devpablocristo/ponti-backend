# dependencies.md — feature-027 BE cleanup & domain purity

## Resumen de grafo

```
001 be-platform-tenancy-refactor ──(débil/cosmética)──> 027 be-cleanup-domain-purity
(capa DTO + mappers ya en develop) ──(fuerte, ya satisfecha)──> 027
027 ── no bloquea a ninguna otra feature
```

## Depende de

| feature/artefacto | tipo | naturaleza | detalle |
|---|---|---|---|
| Capa DTO de report (`internal/report/handler/dto/*.go`) + mappers (`internal/report/usecases/mappers/summary_mappers.go`) + `internal/report/handler.go` que serializa vía DTO | **fuerte** | **YA satisfecha en develop** (verificado PRESENTE) | Es lo que permite quitar los `json:` del dominio sin romper el wire format. El handler hace `dto.BuildFieldCropResponse` / `dto.FromDomainSummaryResults` antes de `c.JSON` |
| `001 be-platform-tenancy-refactor` → crea `internal/report/repository/models/project_info.go` (con `gorm:column:` + `ToDomain()`) | **débil / cosmética** | NO satisfecha en develop (archivo ausente) | El comentario nuevo de `domain.ProjectInfo` apunta a ese archivo. Su ausencia NO rompe build ni runtime; solo deja una referencia documental colgada |

## Bloquea a

- Ninguna feature conocida. 027 es hoja del grafo (cleanup terminal del módulo report).

## Clasificación de certeza

- **Fuertes**: dependencia de la capa DTO (resuelta).
- **Débiles**: orden vs 001 (preferible 001 primero, pero no obligatorio).
- **Inciertas**: ninguna relevante. (La única incertidumbre menor es la NamingStrategy de GORM, ver risks.md — no es una dependencia de feature sino un supuesto técnico, mitigado por el aliasing snake_case del query.)

## Archivos / tipos / config / migraciones / APIs compartidos

- **Archivos compartidos**: ninguno. Los 2 archivos son exclusivos del paquete `internal/report/usecases/domain/`. NO tocan `wire/`, `cmd/api/`, `cmd/config/`, `go.mod`, `go.sum`, `Makefile`, `internal/shared/**`.
- **Tipos compartidos**: los tipos del dominio de report (`ProjectInfo`, `FieldCropMetric`, etc.) son usados por `repository.go`, `usecases.go`, `mappers/` y la capa DTO. El cambio (quitar tags) NO altera nombres ni campos, así que esos consumidores siguen compilando sin cambios.
- **Config**: ninguna.
- **Migraciones**: ninguna.
- **APIs**: los endpoints de report no cambian de forma.

## Cross-repo

- Ninguna dependencia. Solo-BE. En FE: "sin cambios FE".

## Recomendación de orden

1. (Opcional, recomendado) `001 be-platform-tenancy-refactor` — para que exista `repository/models/project_info.go`.
2. `027 be-cleanup-domain-purity` — esta feature.

Si por logística se invierte el orden (027 antes que 001), es aceptable: 027 compila y funciona; solo queda la referencia documental al archivo de models hasta que llegue 001.
