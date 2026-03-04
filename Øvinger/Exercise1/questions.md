Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> Concurrency is about managing multiple task at once and parallelism is about doing many things at the same

What is the difference between a *race condition* and a *data race*? 
> A data race occurs when two or more theads access the same memory location concurrently and at least one access is write and the access are not properly synchronized. A race condition is a bug where a program produces wrong results because multiple threads or processes access and change shared data in an unpredictable order.
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> A scheduler devides the cpu-time and decides which threads gets to run on the CPU, when and how long. 


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> To enable concurrency and parallellism to let applications do multiple things at once to improve the performance.  

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> Fibers are lightweigted units of execution that allow a program to perform multiple tasks concurrently within a single OS thread. 

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> Both easier and harder. It allows programs to handle multiple task and becomes faster, but also introduces complexity and are more prone to errors. 

What do you think is best - *shared variables* or *message passing*?
> It is more simple to message pass since there are no shared state between task and less synchroniziation issues-  


