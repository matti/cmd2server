.PHONY:

test:
	go run cmd/cmd2server/main.go localhost:1234 ping -c 3 google.com &
	nc localhost 1234
