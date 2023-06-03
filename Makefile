web:
	go run . web

worker:
	go run . worker

test: web worker