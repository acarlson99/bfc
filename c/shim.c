#include <stdio.h>

char safe_getchar() {
    char c = getchar();
    if (c < 0) return 0;
    return c;
}
