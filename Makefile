CC = clang

%.ll: %.b bfc
	./bfc $< > $@

%.o: %.ll
	$(CC) -c -o $@ $<

# `make tic-tac-toe` finds b/tic-tac-toe.b and makes it a program
# this is to include `shim.o` for a more consistent definition of `getchar`
%: b/%.o c/shim.o
	$(CC) -o $@ c/shim.o $<

bfc:
	go build .

wc:
xmas: