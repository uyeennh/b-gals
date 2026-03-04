// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>

// Race conditions when two threads changes i at the same time. The threads run simultaneously and the operations can overlap, which will cause some increements or decrements to be lost. 
// This is why the final result might not be 0. 


int i = 0;
pthread_mutex_t lock;

// Note the return type: void*
void* incrementingThreadFunction(void* arg){
    for (int j = 0; j < 1000000; j++){
        pthread_mutex_lock(&lock);
        i++;
        pthread_mutex_unlock(&lock);
    }
    // TODO: increment i 1_000_000 times
    return NULL;
}

// We use mutex because it allows only one function to execute the incrementation og decrementation at a time. This will not work with semaphores.
// This prevents race conditions

void* decrementingThreadFunction(void* arg){
    // TODO: decrement i 1_000_000 times
    for (int j = 0; j < 1000000; j++){
        pthread_mutex_lock(&lock);
        i--;
        pthread_mutex_unlock(&lock);
    }
    return NULL;
}


int main(){
    pthread_t thread1, thread2;
    pthread_mutex_init(&lock, NULL);


    pthread_create(&thread1, NULL, incrementingThreadFunction, NULL);
    pthread_create(&thread2, NULL, decrementingThreadFunction, NULL);

    pthread_join(thread1, NULL);
    pthread_join(thread2, NULL);

    // TODO: 
    // start the two functions as their own threads using `pthread_create`
    // Hint: search the web! Maybe try "pthread_create example"?
    
    // TODO:
    // wait for the two threads to be done before printing the final result
    // Hint: Use `pthread_join`    
    
    printf("The magic number is: %d\n", i);

    pthread_mutex_destroy(&lock);
    return 0;
}
