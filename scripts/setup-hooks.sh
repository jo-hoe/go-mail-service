#!/bin/sh
# Setup script to configure git hooks for this repository

echo "Setting up git hooks..."

# Configure git to use the scripts/git-hooks directory
git config core.hooksPath scripts/git-hooks

echo "✓ Git hooks configured successfully!"
echo "  Hooks directory: scripts/git-hooks"
echo ""
echo "The following hooks are now active:"
echo "  - pre-commit: Runs 'go fmt ./...' before each commit"
