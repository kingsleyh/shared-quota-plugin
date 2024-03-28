#!/bin/bash
# This script serves a helper script to run the Tyk bundler tool to create a production-ready plugin bundle.
set -euo pipefail;
echo "Building plugin bundle...";

# Copy custom plugin to bundle directory
cp /opt/tyk-gateway/middleware/SharedQuotaPlugin*.so /opt/tyk-gateway/bundle/;

# Run bundler tool in bundle directory
cd /opt/tyk-gateway/bundle && /opt/tyk-gateway/tyk bundle build -y;

# Cleanup
rm /opt/tyk-gateway/bundle/SharedQuotaPlugin*.so;

# Exit
echo "Done building plugin bundle.";
exit 0;
