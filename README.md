# SS-project

## WORKFLOW

# 1. Always start from an up-to-date main
git checkout main
git pull origin main

# 2. Create a branch for your feature/fix
git checkout -b feature/your-feature-name

# 3. Work, then commit often
git add .
git commit -m "short description of what you did"

# 4. Push your branch
git push origin feature/your-feature-name

# 5. Open a Pull Request on GitHub, request a review
# 6. After approval → merge → delete the branch


# 7. If main gets new commits while you're working on your branch, sync up:
git checkout your-branch
git rebase main


## Branch naming conventions

Type            Example
New feature     feature/login-page
Bug fix         fix/crash-on-startup
Hotfix          hotfix/security-patch


## Good commit messages

# Bad
git commit -m "stuff"
git commit -m "fix"

# Good
git commit -m "add user authentication"
git commit -m "fix null pointer in login form"
