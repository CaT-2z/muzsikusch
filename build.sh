#!/bin/bash
# This will build the binary and add the capability to bind to port 443
go build && sudo setcap 'cap_net_bind_service=+ep' muzsikusch