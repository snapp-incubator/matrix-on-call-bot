package config

const defaultConfig = `
matrix:
  url: "https://example.com"
  userID: "@some_bot:example.com"
  token: "secret"
  display-name: "user"

database:
  driver: mysql
  host: "localhost"
  port: 33060
  db_name: "on-call"
  username: "on-call"
  password: "secret"
  timeout: "30s"
  read_timeout: "3s"
  write_timeout: "1s"
  connection_lifetime: "10m"
  max_open_connections: 10
  max_idle_connections: 5
`
