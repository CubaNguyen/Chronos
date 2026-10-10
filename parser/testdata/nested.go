package nested

func Outer() {
    Inner()
    func() {
        DeepNested()
    }()
}

func Inner() {}
func DeepNested() {}