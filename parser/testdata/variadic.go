package variadic

import "fmt"

func Sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    fmt.Println("sum computed")
    return total
}