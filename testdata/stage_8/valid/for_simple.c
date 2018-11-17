int main() {
    int a = 0;
    for (int i = 0; i < 10; i = i + 1) {
        for (int j = 0; j < 10; j = j + 1) {
            a = a + 1;
        }
    }
    return a;
}