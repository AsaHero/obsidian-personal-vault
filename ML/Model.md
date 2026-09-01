## Introduction
---
In this section, we explore what a **model** is, how we can evaluate if a model is learning, and how to determine whether it is a **good** or **bad** model.

 *Dataset: Supervised Learning*

|**Pregnancies**|**Glucose**|**BloodPressure**|**SkinThickness**|**Insulin**|**BMI**|**Age**|**Outcome**|
|---|---|---|---|---|---|---|---|
|6|148|72|35|0|33.6|50|1|
|1|85|66|29|0|26.6|31|0|
|8|183|64|0|0|23.3|32|1|
|1|89|66|23|94|28.1|21|0|
|0|137|40|35|168|43.1|33|1|
|5|116|74|0|0|25.6|30|0|
|3|78|50|32|88|31.0|26|1|
|10|115|0|0|0|35.3|29|0|
|2|197|70|45|543|30.5|53|1|
|8|125|96|0|0|0.0|54|1|
|4|110|92|0|0|37.6|30|0|
|10|168|74|0|0|38.0|34|1|

- **Each Row**: Represents a different sample in the dataset.
- **Each Column (except Outcome)**: Represents a different feature.
- **Outcome Column**: The **output label**, which indicates whether the person has diabetes (`1`) or not (`0`).

## Key Concepts
---
#### 1. Features Vector:

- Each row in the dataset (excluding the `Outcome` column) is a **features vector**.
- Example: `[6, 148, 72, 35, 0, 33.6, 50]` is a features vector.
#### 2. Output Label:

 - The `Outcome` column represents the label for each features vector. Example: `1` (diabetes) or `0` (no diabetes).
#### 3. Features Matrix (X):

- The **entire table excluding the Outcome column** is called the **features matrix (X)**.
#### 4. Output Labels Vector (Y): 

- The `Outcome` column is called the **output labels vector (Y)**.

## Model Training Workflow
---
1. **Input to the Model**:
    
    - Each row (features vector) from the features matrix `X` is fed into the model.
2. **Prediction**:
    
    - The model predicts an output `y-predicted`.
3. **Comparison**:
    
    - Compare `y-predicted` (model's output) to `y-actual` (true value from labels vector `Y`).
4. **Adjustments**:
    
    - The model adjusts its parameters to get closer to the true value (`y-actual`) in the next iteration.
5. **Training Loop**:
    
    - This process of feeding data, making predictions, and adjusting the model is repeated in a **loop**, known as **training** the model


## Dataset Splits
---
In supervised learning, the dataset is divided into three subsets to train, validate, and test the model:

#### 1. Training Dataset:
    
- Used to train the model.
- The model generates an output vector, which is compared to the actual output vector to calculate the **Loss**.
- The loss is used to make **adjustments** in the model to improve its performance.
#### 2. Validation Dataset:
    
- Used to calculate **Loss** after training to evaluate how well the model performs on unseen data.    
- **No adjustments** are made to the model based on the validation dataset.
- A lower validation loss indicates that the model is performing better.
#### 3. Test Dataset:
    
 - Used to assess the final performance of the chosen model with the least loss value (from training and validation).    
 - Ensures that the model generalizes well to **new, unseen data**.   
 - Final statistics (e.g., accuracy) are reported based on the test dataset.


## Metrics of Performance
---
### What is Loss?

Loss measures the difference between the model's predictions (`Y_predicted`) and the actual values (`Y_actual`).

- **The smaller the loss, the better the model is performing.**
### Loss Functions

Loss functions are used to calculate the loss value. Common examples include:

1. **L1 Loss** (Mean Absolute Error):
    
    - Formula:  
        `loss = sum(|Y_actual - Y_predicted|)`
    - Measures the absolute differences between predictions and true values.
2. **L2 Loss** (Mean Squared Error):
    
    - Formula:  
        `loss = sum((Y_actual - Y_predicted)^2)`
    - Penalizes larger errors more heavily than L1 Loss.
3. **Binary Cross-Entropy Loss**:
    
    - Used for classification problems, especially **binary classification** (e.g., spam vs. not spam).
    - The loss decreases as the model's confidence in correct predictions increases.
### What is Accuracy?

Accuracy measures the proportion of correct predictions out of total predictions:

- Formula:  
    `Accuracy = (Number of Correct Predictions) / (Total Predictions)`

Example:  
If the model makes **3 correct predictions out of 4**, the accuracy is:  
`3/4 = 75%`





