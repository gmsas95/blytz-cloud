# Git Branching Strategy

## Branch Overview

This project uses a **Git Flow** branching strategy with three main branches:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    main     │     │   staging   │     │   feature   │
│  (Production)│◄────│  (Staging)  │◄────│   branches  │
│   blytz.cloud│     │staging.blytz│     │             │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Branch Descriptions

### `main` - Production
- **Purpose**: Production-ready code
- **Deployment**: Live production environment
- **URL**: https://blytz.cloud
- **Protection**: Direct pushes disabled, requires PR from staging
- **Stability**: Must always be stable and deployable

### `staging` - Pre-Production  
- **Purpose**: Testing and QA before production
- **Deployment**: Staging environment
- **URL**: https://staging.blytz.cloud (or localhost:3000 for local)
- **Protection**: Requires PR from feature branches
- **Stability**: Should be stable, but can have bugs

### `feature/*` - Development
- **Purpose**: Individual feature development
- **Naming**: `feature/agent-marketplace`, `feature/payment-stripe`, etc.
- **Deployment**: Local development only
- **Lifecycle**: Created from staging, merged back to staging

## Workflow

### 1. Starting a New Feature

```bash
# Ensure you're on staging and it's up to date
git checkout staging
git pull origin staging

# Create feature branch
git checkout -b feature/your-feature-name

# Make changes, commit, push
git add .
git commit -m "Add: your feature description"
git push -u origin feature/your-feature-name
```

### 2. Merging to Staging

```bash
# Create PR on GitHub from feature/* to staging
# Or manually:
git checkout staging
git merge feature/your-feature-name
git push origin staging
```

### 3. Deploying to Staging

```bash
# On staging server
git checkout staging
git pull origin staging

# Rebuild
npm install && npm run build  # Frontend
go build -o blytz ./cmd/server  # Backend

# Restart services
sudo systemctl restart blytz
sudo systemctl restart blytz-frontend
```

### 4. Promoting to Production

```bash
# Create PR on GitHub from staging to main
# After review and approval:
git checkout main
git merge staging
git push origin main

# Deploy to production
# (Follow production deployment guide)
```

## Branch Protection Rules

### Main Branch
- [x] Require pull request reviews before merging
- [x] Require status checks to pass
- [x] Require branches to be up to date before merging
- [x] Include administrators
- [x] Restrict pushes that create files larger than 100MB

### Staging Branch  
- [x] Require pull request reviews before merging
- [x] Allow force pushes (for emergencies)
- [x] Allow deletions

## Commit Message Convention

Format: `<type>: <subject>`

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, semicolons, etc)
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Build process or auxiliary tool changes

Examples:
```
feat: add agent marketplace page
fix: resolve port allocation race condition  
docs: update deployment guide
style: format dashboard components
refactor: extract sidebar into component
test: add billing page tests
chore: update dependencies
```

## Deployment Environments

| Branch | Environment | URL | Purpose |
|--------|-------------|-----|---------|
| `main` | Production | https://blytz.cloud | Live users |
| `staging` | Staging | https://staging.blytz.cloud | Testing |
| `feature/*` | Local | http://localhost:3000 | Development |

## Environment Variables by Branch

### Production (main)
```bash
BASE_DOMAIN=blytz.cloud
DATABASE_PATH=/opt/blytz/data/production.db
LOG_LEVEL=warn
```

### Staging (staging)  
```bash
BASE_DOMAIN=staging.blytz.cloud
DATABASE_PATH=/opt/blytz/data/staging.db
LOG_LEVEL=debug
```

### Development (feature/*)
```bash
BASE_DOMAIN=localhost
DATABASE_PATH=./tmp/dev.db
LOG_LEVEL=debug
```

## Testing Strategy

### Before Merging to Staging
- [ ] Unit tests pass
- [ ] Manual testing on local
- [ ] Code review completed
- [ ] No console errors

### Before Merging to Main
- [ ] All staging tests pass
- [ ] Integration tests pass
- [ ] Performance testing
- [ ] Security review
- [ ] Documentation updated

## Hotfix Workflow

For urgent production fixes:

```bash
# Create hotfix from main
git checkout main
git pull origin main
git checkout -b hotfix/critical-bug

# Fix, commit, push
git add .
git commit -m "fix: resolve critical bug"
git push -u origin hotfix/critical-bug

# Create PR to main AND staging
git checkout staging
git merge hotfix/critical-bug
git push origin staging
```

## Release Process

### Version Numbering
- Format: `v{major}.{minor}.{patch}`
- Example: `v1.2.3`

### Creating a Release

```bash
# On main branch
git checkout main
git pull origin main

# Tag release
git tag -a v1.2.3 -m "Release v1.2.3 - Agent Marketplace"
git push origin v1.2.3

# Create GitHub release with notes
```

### Release Checklist
- [ ] Version bumped in code
- [ ] CHANGELOG.md updated
- [ ] Migration scripts tested
- [ ] Documentation updated
- [ ] Release notes prepared

## Collaboration Guidelines

### 1. Always Work on Feature Branches
```bash
# ❌ Don't do this
git checkout staging
git commit -m "quick fix"

# ✅ Do this
git checkout -b feature/quick-fix
git commit -m "fix: resolve edge case"
```

### 2. Keep Feature Branches Short-Lived
- Create → Develop → PR → Merge (within 1 week ideally)
- Long-lived branches = merge conflicts

### 3. Regularly Sync with Staging
```bash
git checkout feature/my-feature
git fetch origin
git rebase origin/staging
```

### 4. Meaningful Commit Messages
```bash
# ❌ Bad
git commit -m "fix"

# ✅ Good
git commit -m "fix: prevent duplicate port allocation in concurrent requests"
```

## Quick Reference

### Switch to Branch
```bash
git checkout staging
git checkout main
git checkout feature/my-feature
```

### Create Feature Branch
```bash
git checkout staging
git pull origin staging
git checkout -b feature/new-feature
```

### Merge Feature to Staging
```bash
git checkout staging
git merge feature/new-feature
git push origin staging
```

### Promote Staging to Production
```bash
git checkout main
git merge staging
git push origin main
```

### Delete Merged Feature Branch
```bash
git branch -d feature/new-feature      # Local
git push origin --delete feature/new-feature  # Remote
```

## Troubleshooting

### Merge Conflicts
```bash
git checkout staging
git merge feature/my-feature

# If conflicts:
# 1. Edit files to resolve
# 2. git add .
# 3. git commit -m "merge: resolve conflicts"
```

### Accidental Commit to Wrong Branch
```bash
# Undo last commit, keep changes
git reset --soft HEAD~1

# Switch to correct branch
git checkout feature/correct-branch

# Re-commit
git add .
git commit -m "feat: your feature"
```

### Staging is Behind Main
```bash
git checkout staging
git merge main
git push origin staging
```

---

**Current Branches:**
- `main` → Production (blytz.cloud)
- `staging` → Staging environment
- Create feature branches from `staging`

**Workflow:** feature → staging → main
