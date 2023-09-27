
bin:
	mkdir bin

bin/sensor: bin mock_sensor/*
	go build -o bin/sensor ./mock_sensor

.PHONY: clean
clean:
	pkill sensor || true
	docker rm -f collector || true
