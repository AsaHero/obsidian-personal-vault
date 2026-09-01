[[Equations]] have maximum and minimum points on the [[Two Variable Equations#Cartesian Plane|Cartesian Plane]]. This means in the set of solutions of [[equations]] we could have a maximum or minimum numbers where the pairs are increase or decrease respectively. 

## Absolute minimum
___
Let's see if the graph of an equation has a **lowest** point: 
$$y=2x^2+3$$
```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-10,10,-10,10]
disableZoom: true
grid: true
---
y=2x^2+3
```

The lowest point of y is 3, it cannot be lower than this, which is an **absolute minimum** of y-value in equation.

> [!Absolute minimum]
> The **absolute minimum** value of a variable in an equation is the smallest it can be in a solution.

## Absolute maximum
___
Let's see how **large** y can be in another equation:
$$y = 6 - |x - 4|$$
```functionplot
---
title: 
xLabel: 
yLabel: 
bounds: [-10,10,-10,10]
disableZoom: true
grid: true
---
y=6-abs(x-4)
```

The largest value y-value will have is 6, which is an **absolute maximum** of y-value in equation.

>[!Absolute maximum]
>The **absolute maximum** value of a variable in an equation is the largest it can be in a solution.

## Local minimum and Local maximum
---
Let's see how maximum and minimum values relate to the way variables **change** in an equation. Let's consider this cubic equation for sake of explanation: 
$$y = x^2(x - 3)$$
```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-2,5,-5,5]
disableZoom: true
grid: true
---
y=x^2(x-3)
```

When $0<x<2$ , y **decrease** as x increases. A (0; 0) is a point when y changes from increasing to decreasing, and (2; -4) changes vice verse.   

We say y reaches a **local** minimum (or maximum) when the point is lower (or higher) than any nearby points. Subsequently, (0; 0) is local maximum related to the point (2; -4), which is local minimum related to the (0; 0).

>[!Note !]
>When a graph's behavior changes from increasing to decreasing or vice versa,  y can reach a **local minimum** or **local maximum**. In some cases local minimum and local maximum can be absolute minimum and absolute maximum respectively. 
>


