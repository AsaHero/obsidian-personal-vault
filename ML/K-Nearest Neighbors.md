## Intro

K-Nearest Neighbor (KNN) is a simple algorithm used for **classification** and **regression** tasks. It relies on the proximity of data points to make predictions, without learning an explicit model.

## How KNN Works

1. **K Value**: Determines how many neighbors to consider. Example: For `K=3`, the algorithm looks at the 3 nearest data points.
2. **Distance Metric**: Calculates the "closeness" of points using metrics like **Euclidean distance**: 
$$d(p,q)=\sqrt{\sum (p_i - q_i)^2}​$$
3. **Prediction**:
    - **Classification**: Majority vote among the K neighbors.
    - **Regression**: Average value of the K neighbors.

## Example: Predicting Car Ownership

#### Problem:

Use KNN to predict whether a person owns a **car** or **no car** based on:

- **Number of kids** (X-axis).
- **Amount of salary** (Y-axis).

#### Visualization:

```vega-lite
{
  "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
  "data": {
    "values": [
      {"kids": 1, "salary": 50000, "class": "Car"},
      {"kids": 3, "salary": 60000, "class": "Car"},
      {"kids": 1, "salary": 55000, "class": "Car"},
	  {"kids": 2, "salary": 60000, "class": "Car"},
      {"kids": 1, "salary": 20000, "class": "Car"},
      {"kids": 4, "salary": 30000, "class": "No Car"},
      {"kids": 5, "salary": 25000, "class": "No Car"},
      {"kids": 0, "salary": 35000, "class": "No Car"},
      {"kids": 2, "salary": 30000, "class": "No Car"},
      {"kids": 0, "salary": 25000, "class": "No Car"},
      {"kids": 1, "salary": 35000, "class": "No Car"},
      {"kids": 2, "salary": 50000, "class": "New Point"}
    ]
  },
  "mark": "point",
  "encoding": {
    "x": {"field": "kids", "type": "quantitative", "axis": {"title": "Number of Kids"}},
    "y": {"field": "salary", "type": "quantitative", "axis": {"title": "Salary"}},
    "color": {"field": "class", "type": "nominal", "scale": {"scheme": "category10"}},
    "shape": {"field": "class", "type": "nominal"}
  }
}

```
#### Explanation:

- The new point (2,50k) represents someone with 3 kids and a salary of $50,000.
- Using K=3, the closest neighbors are:
    1. (1,50k) (Car)
    2. (1,55k) (Car)
    3. (2,60k) (Car)
 
Majority vote: **Car** → Prediction: **Car**.