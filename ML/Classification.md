## Overview 

As we know classification can be divided into two: multi-task classification and binary classification. 

There are several model designs for classification tasks: 
1. [[K-Nearest Neighbors|KNN (K-Nearest Neighbors)]] - simplest model, to classify according to nearest features on the **n-dimensional Euclidean space**.


## Key Metrics

![[Pasted image 20241125222235.png]]

This diagram visually explains classification results and their connection to the performance metrics, precision and recall, in a simple and intuitive way. The rectangle represents the entire dataset, split into two parts: the left half contains relevant elements (ground truth positives), while the right half contains non-relevant elements (ground truth negatives). A circle inside the rectangle represents the model's predictions, specifically the elements it retrieves, and is divided into two halves as well. The left side of the circle corresponds to true positives, the relevant items that the model correctly retrieved. The right side of the circle corresponds to false positives, the non-relevant items incorrectly retrieved as relevant by the model. Outside the circle, the left portion of the rectangle represents false negatives, the relevant items that the model failed to retrieve. The right portion of the rectangle outside the circle represents true negatives, the non-relevant items that the model correctly predicted as non-relevant. This structure effectively connects the classification outcomes to two critical metrics: precision and recall. Precision reflects the accuracy of the retrieved items, calculated as the ratio of true positives to all items within the circle. Recall measures the completeness of the model's retrieval, calculated as the ratio of true positives to all relevant items in the rectangle. This diagram elegantly highlights the trade-offs and alignment between the model's predictions and the ground truth, offering a clear way to interpret classification results. 

Here is the confusion matrix that corresponds to the scenario described, summarizing the classification results:

|**Actual / Predicted**|**Positive** (Retrieved)|**Negative** (Not Retrieved)|
|---|---|---|
|**Positive** (Relevant)|True Positives (TP)|False Negatives (FN)|
|**Negative** (Non-Relevant)|False Positives (FP)|True Negatives (TN)|



### 1. Accuracy
---

- **Definition:** The ratio of correctly predicted instances to the total instances.
- **Formula:** $$Accuracy = \frac{TP + TN}{TP + TN + FP + FN}$$
- **Best For:** Balanced datasets with equal class distribution.
- **Limitations:** Misleading on imbalanced datasets (e.g., a model predicting all instances as the majority class may achieve high accuracy).

### 2. Precision
---

- **Definition:** The ratio of correctly predicted positive observations to the total predicted positives.
- **Formula:** $$Precision = \frac{TP}{TP + FP}$$
- **Use Case:** Important when minimizing false positives is critical (e.g., spam detection).


### 3. Recall (Sensitivity or True Positive Rate)
---

- **Definition:** The ratio of correctly predicted positive observations to all actual positives.
- **Formula:** $$Recall = \frac{TP}{TP + FN}$$
- **Use Case:** Critical when minimizing false negatives is important (e.g., medical diagnosis).

### 4. F1 Score
---

- **Definition:** The harmonic mean of Precision and Recall.
- **Formula:** $$F1 = 2 \cdot \frac{Precision \cdot Recall}{Precision + Recall}​$$
- **Use Case:** Balanced metric for uneven class distributions, combining Precision and Recall into a single measure.

### 5. Specificity (True Negative Rate)
---

- **Definition:** The ratio of correctly predicted negative observations to all actual negatives.
- **Formula:** $$Specificity = \frac{TN}{TN + FP}$$​
- **Use Case:** Useful when distinguishing true negatives is important (e.g., fraud detection).


### 6. ROC-AUC (Receiver Operating Characteristic - Area Under Curve)
---

- **Definition:** Measures the ability of a classifier to distinguish between classes.
- **Explanation:**
    - **ROC Curve:** Plots True Positive Rate (Recall) against False Positive Rate.
    - **AUC:** Represents the area under the ROC curve, ranging from 0.5 (random) to 1 (perfect).
- **Use Case:** Evaluates model performance across different classification thresholds.

### 7. Logarithmic Loss (Log Loss)
---
- **Definition:** Measures the uncertainty of predictions by comparing predicted probabilities to actual labels.
- **Formula:** $$Log Loss = -\frac{1}{N} \sum_{i=1}^{N} \left[ y_i \log(p_i) + (1 - y_i) \log(1 - p_i) \right]$$
- **Use Case:** Appropriate for probabilistic models where confidence in predictions matters.
