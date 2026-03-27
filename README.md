# SS-project

> Short description.

---

## 📁 Project Structure
```
SS-project/
├── src/
├── tests/
└── README.md
```


---

## 🔄 Workflow

Always follow these steps — no direct pushes to `main`!

### 1. Start from an up-to-date main
```bash
git checkout main
git pull origin main
```

### 2. Create a branch for your feature/fix
```bash
git checkout -b feature/your-feature-name
```

### 3. Work, then commit often
```bash
git add .
git commit -m "short description of what you did"
```

### 4. Push your branch
```bash
git push origin feature/your-feature-name
```

### 5. Open a Pull Request on GitHub and request a review

### 6. After approval → merge → delete the branch

### 7. Keep your branch up to date with main
```bash
git checkout your-branch
git rebase main
```

---

## 🌿 Branch Naming Conventions

| Type        | Example                  |
|-------------|--------------------------|
| New feature | `feature/login-page`     |
| Bug fix     | `fix/crash-on-startup`   |
| Hotfix      | `hotfix/security-patch`  |

---

## ✅ Good Commit Messages
```bash
# ❌ Bad
git commit -m "stuff"
git commit -m "fix"

# ✅ Good
git commit -m "add user authentication"
git commit -m "fix null pointer in login form"
```

---

## 🔒 Branch Rules

- No direct pushes to `main`
- All changes go through a **Pull Request**
- At least **1 approval** required before merging
