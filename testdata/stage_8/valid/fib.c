int main() {
    
    int n = 20;
    int f1 = 0;
    int f2 = 1;
    int fi;
 
    for (int i = 2 ; i <= n ; i = i + 1) {
        fi = f1 + f2;
        f1 = f2;
        f2 = fi;
    }

    return fi;
}