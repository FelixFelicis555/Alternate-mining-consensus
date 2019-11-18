s = "Hello World"
for i in range(50):
    f = open("text{:02}.txt".format(i+1),"w+")
    for i in range(9):
        f.write(s)
        f.write("\n")
    f.write("s")
    f.close()