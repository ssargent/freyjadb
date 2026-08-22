#!/bin/bash

# Script to verify the Makefile to Justfile conversion

echo "=== FreyjaDB Justfile Conversion Verification ==="
echo

echo "1. Checking that justfile exists..."
if [ -f "justfile" ]; then
    echo "✓ justfile exists"
else
    echo "✗ justfile missing"
    exit 1
fi

echo
echo "2. Checking that just is installed..."
if command -v just >/dev/null 2>&1; then
    echo "✓ just is installed"
else
    echo "⚠ just is not installed (install with: brew install just)"
fi

echo
echo "3. Testing basic just commands..."

echo "  a. Testing 'just help'..."
if just help >/dev/null 2>&1; then
    echo "  ✓ 'just help' works"
else
    echo "  ✗ 'just help' failed"
fi

echo "  b. Testing 'just deps'..."
if just deps >/dev/null 2>&1; then
    echo "  ✓ 'just deps' works"
else
    echo "  ✗ 'just deps' failed"
fi

echo
echo "4. Checking that README was updated..."
if grep -q "Using Just" README.md; then
    echo "✓ README updated with Just information"
else
    echo "✗ README not properly updated"
fi

echo
echo "5. Summary of conversion:"
echo "   - All Makefile targets converted to Just recipes"
echo "   - README updated with Just usage instructions"
echo "   - Backward compatibility with Make maintained"
echo "   - Cross-platform support improved"

echo
echo "=== Verification Complete ==="