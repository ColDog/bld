#!/bin/bash

data=$( cat pkg/builder/schema.json | tr -d ' \t\n\r\f' )

cat > pkg/builder/schema.go <<EOF
package builder

const schema = \`${data}\`
EOF
