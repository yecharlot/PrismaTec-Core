# Cómo llegar el código a GitHub: `PrismaTec-Core`

## Dónde está el código ahora

En esta máquina de trabajo el repositorio **local** es:

```text
/home/workdir/artifacts/PrismaTec-Core
```

Ese es el árbol fuente de PrismaTec Core (Fases 0–5).  
**Aún no está automáticamente en GitHub** hasta que tú (o una IA con credenciales) lo publiques.

## URL objetivo

```text
https://github.com/yecharlot/PrismaTec-Core
```

(Coincide con el manifiesto: `github.com/yecharlot/PrismaTec-Core`)

## Ruta paso a paso (en tu ordenador o donde tengas `gh` / git + token)

### 1) Copia o abre el directorio del Core

Si trabajas en este entorno:

```bash
cd /home/workdir/artifacts/PrismaTec-Core
```

Si lo descargas desde otro sitio, usa la carpeta donde esté el código.

### 2) Inicializa git (si aún no hay `.git`)

```bash
cd /home/workdir/artifacts/PrismaTec-Core
git init
git add .
git status
git commit -m "PrismaTec Core: phases 0-5 (boot, organism, memory/cid, pulse, AIP v1)"
```

### 3) Crea el repo vacío en GitHub

**Opción A — GitHub web**

1. Entra en https://github.com/new  
2. Owner: **yecharlot**  
3. Repository name: **PrismaTec-Core**  
4. Public (o Private)  
5. **No** marques “Add README” (ya tenemos código)  
6. Create repository  

**Opción B — GitHub CLI**

```bash
gh auth login
gh repo create yecharlot/PrismaTec-Core --public --source=. --remote=origin --push
```

### 4) Conecta remote y sube (si usaste la web)

```bash
git branch -M main
git remote add origin https://github.com/yecharlot/PrismaTec-Core.git
git push -u origin main
```

Si GitHub te pide autenticación: Personal Access Token (classic) con scope `repo`, o SSH:

```bash
git remote add origin git@github.com:yecharlot/PrismaTec-Core.git
git push -u origin main
```

### 5) Comprueba

Abre: https://github.com/yecharlot/PrismaTec-Core  

Debes ver `cmd/`, `core/`, `api/`, `docs/`, `README.md`, etc.

## Relación con los otros repos (no los borres)

| Repo | URL | Rol |
|------|-----|-----|
| AlsetOS | https://github.com/yecharlot/AlsetOS | Origen runtime |
| PrismaTec | https://github.com/yecharlot/PrismaTec | Origen intelligence/edge |
| Alset-JS-Runtime | https://github.com/yecharlot/Alset-JS-Runtime | Origen UI/AIP client |
| **PrismaTec-Core** | https://github.com/yecharlot/PrismaTec-Core | **Nuevo** — unificación |

## Nota para IAs en este sandbox

Este entorno puede **no** tener token de GitHub de `yecharlot`.  
Puede preparar el commit local; el **push** lo hace el dueño de la cuenta o un agente con `GH_TOKEN` / SSH configurado.

## Variables útiles al desarrollar

```bash
export GOPROXY=https://proxy.golang.org,direct
export GOTOOLCHAIN=go1.25.0
export PRISMATEC_DATA_DIR=$HOME/.prismatec/node
export PRISMATEC_AIP_ADDR=:8080
```
