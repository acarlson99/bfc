CC = clang

LLVM_TARGET_ARCH = $(shell llvm-config --host-target)

B_DIR = b/
B_NAMES = $(wildcard $(B_DIR)*.b)
B_TARGETS = $(B_NAMES:$(B_DIR)%.b=%)
TARGETS = $(B_TARGETS) bfc

%.ll: %.b bfc
	./bfc -o $@ -arch $(LLVM_TARGET_ARCH) $<

%.o: %.ll
	$(CC) -c -o $@ $<

# `make tic-tac-toe` finds b/tic-tac-toe.b and makes it a program
# this is to include `shim.o` for a more consistent definition of `getchar`
%: $(B_DIR)%.o c/shim.o
	$(CC) -o $@ c/shim.o $<

bfc:
	go build .

all: $(TARGETS)

clean:
	rm -f $(TARGETS)

re: clean all
