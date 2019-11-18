def main():
    avg = [0,0,0,0]
    with open('result_5KB.txt','r') as f:
        for line in f:
            x = line.split(' ')
            block = int(x[2])
            du = x[4]
            m = du.split('m')
            time = 0
            if len(m) == 1 :
                time = float(m[0].split('s')[0])
            else:
                time = 60*int(m[0]) + float(m[1].split('s')[0])
            if block >= 1 and block <= 50 :
                avg[0] += time 
            if block >= 50 and block <= 100 :
                avg[1] += time 
            if block >= 100 and block <= 150 :
                avg[2] += time 
            if block >= 150 and block <= 200 :
                avg[3] += time 
    final = 0 
    for i in range(len(avg)):
        print('Block: ',i*50,'to',(i+1)*50,'::',avg[i]/50)
        final += avg[i] 
    print("Final Avg: ",final/200)

if __name__ == '__main__':
    main()