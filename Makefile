CC = clang

LLVM_TARGET_ARCH = $(shell llvm-config --host-target)

BF_SRC_DIR = b/
BF_SRC_NAMES = $(wildcard $(BF_SRC_DIR)*.b)
BF_SRC_TARGETS = $(BF_SRC_NAMES:$(BF_SRC_DIR)%.b=%)
SHIM_O = c/shim.o
BIN_DIR = bin/
BFC = $(BIN_DIR)bfc
TARGETS = $(addprefix $(BIN_DIR), $(BF_SRC_TARGETS)) $(BFC)

all: $(BIN_DIR) $(TARGETS)

clean:
	rm -f $(TARGETS)
	rmdir $(BIN_DIR)

re: clean all

%.ll: %.b $(BFC)
	$(BFC)  -o $@ --arch=$(LLVM_TARGET_ARCH) $<

%.o: %.ll
	$(CC)    -c -o $@ $<

$(BIN_DIR):
	echo $(TARGETS)
	mkdir -p $(BIN_DIR)

# `make tic-tac-toe` finds b/tic-tac-toe.b and makes it a program
# this is to include `shim.o` for a more consistent definition of `getchar`
$(BIN_DIR)%: $(BF_SRC_DIR)%.o $(SHIM_O)
	$(CC)    -o $@ $(SHIM_O) $<

$(BIN_DIR)%: %.go
	go build -o $@ $<
