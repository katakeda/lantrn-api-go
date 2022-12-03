CMD = go
FLAGS = -v -x
TARGET = ./lantrn-api-go
SRC = ./main.go
RM = /bin/rm -f

.PHONY: run
run:
	$(CMD) run $(SRC)

.PHONY: build
build:
	$(CMD) build -o $(TARGET) $(FLAGS) $(SRC)

.PHONY: clean
clean:
	$(RM) $(TARGET)