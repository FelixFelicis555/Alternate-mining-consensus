s = "Hello World"
for i in range(10):
    f = open("text{:02}.txt".format(i+1),"w+")
    for i in range(10):
        f.write(s)
        f.write("\n")
    f.close()