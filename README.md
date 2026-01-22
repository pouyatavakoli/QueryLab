# how to use
1. create db user `querylab_admin`
2. create db user `querylab_sandbox`
```sql
sudo -u postgres psql -c "CREATE USER querylab_sandbox WITH PASSWORD 'sandbox-strong-password';"
```
3. create db `querylab`
4. run `go run cmd/server/main.go`
5. write the db base in init.sql




# TODO
- show only table header when table empty (currently shows nothing)
- show loading process to user
- 


