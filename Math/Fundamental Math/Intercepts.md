In [[Two Variable Equations#Cartesian Plane|Cartesian Plane]], graphs of our [[equations]] can cross x-axis and/or y-axis. The points where a graph meets the axes are called **intercepts**. 

> [!Note !]
> An **x-intercept** is where a graph meets the x-axis.
> A **y-intercept** is where a graph meets the y-axis.

We find x-intercepts by making y=0, and y-intercepts by making x=0. Let's see some examples:

$$10x+20y=60$$
Let's find x-intercept. This means we should solve the equation by making y=0.
$$\displaylines{10x+20\times0 = 60\\10x = 60\\x=6}$$
Let's find y-intercept. This means we should solve the equation by making x=0.
$$\displaylines{10\times0+20y = 60\\20y = 60\\y=3}$$
The results, the equation solution graph intercepts the x-axis when x=6, and y-axis when y=3.

```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-10,10,-10,10]
disableZoom: true
grid: true
---
y=(60-10x)/20
```

A equation graph can have multiple x or y intercepts. For example quadratic equations:
$$x^2+ 2y = 10$$
```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-10,10,-10,10]
disableZoom: true
grid: true
---
y=(x^2 - 6)/2
```
