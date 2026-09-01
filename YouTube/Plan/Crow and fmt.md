fmt installation:
```bash
sudo apt install libfmt-dev
```

Crow installation:
https://github.com/CrowCpp/Crow - release packages  

CMake configuration:
```cmake

find_package(Crow REQUIRED)
find_package(fmt REQUIRED)
target_link_libraries(${PROJECT_NAME}
    PRIVATE
        Crow::Crow
        fmt::fmt
)
```


