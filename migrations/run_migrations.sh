echo $DATABASE_URL
alias run="GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING=$DATABASE_URL goose -dir $PATH_MIGRATIONS_DIR up"
for i in 1 2 3 4 5; do run && break || sleep 15; done

