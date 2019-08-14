#!/bin/bash

set -x

go vet ./cmd/...
go vet ./configs/...
go vet ./dbms/...
go vet ./global/...
go vet ./service/...