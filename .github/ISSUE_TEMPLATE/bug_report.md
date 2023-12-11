---
name: Bug report
about: Create a report to help us improve
title: "[BUG]"
labels: bug
assignees: ''

---

**Checklist**
- [ ] I have used the latest version of terraform (if not please provide a version number).
- [ ] I have used the latest version of this terraform provider (if not please provide a version number).
- [ ] I have provided a sample .tf configuration under "Additional context".
- [ ] I have answered the following question: Is this issue reproducible in OT&E or only applies to live?

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Additional context**
Add any other context about the problem here.

Sample .tf configuration:
```terraform
terraform {
  required_providers {
    inwx = {
      source = "inwx/inwx"
      version = ">= 1.0.0"
    }
  }
}

provider "inwx" {
  api_url = var.api_url
  tan = var.tan
  username = var.username
}
```
