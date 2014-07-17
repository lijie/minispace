#ifndef	__MP_FIFO_H__
#define	__MP_FIFO_H__

#include <stdint.h>

typedef struct fifo {
	int gt;
	int pt;
	size_t size;
	uintptr_t data[0];
} fifo_t;

static inline fifo_t * fifo_new(size_t num) {
	size_t size = sizeof(fifo_t) + sizeof(uintptr_t) * num;
	fifo_t *f = (fifo_t *)malloc(size);
	f->gt = 0;
	f->pt = 0;
	f->size = num;
	return f;
}

static inline int fifo_empty(fifo_t *f) {
	return f->gt == f->pt;
}

static inline int fifo_full(fifo_t *f) {
	return ((f->pt + 1) & (f->size - 1)) == f->gt;
}

static inline void fifo_put(fifo_t *f, uintptr_t d) {
	int pt = f->pt;
	f->data[pt] = d;
	f->pt = (pt + 1) & (f->size - 1);
}

static inline uintptr_t fifo_get(fifo_t *f) {
	int gt = f->gt;
	uintptr_t d = f->data[gt];
	f->gt = (gt + 1) & (f->size - 1);
	return d;
}

#endif
