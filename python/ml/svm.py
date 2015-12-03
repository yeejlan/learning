import struct

tdxdaybase = 'C:/new_gtja_v6/Vipdoc/sh/lday'
stockcode = 'sh000001'


def file_get_contents(filename, mode='r'):
    try:
        f = open(filename, mode)
    except IOError:
        exc_type, exc_value = sys.exc_info()[:2]
        errmsg = '{}: {}'.format(exc_type.__name__, exc_value)
        #output(errmsg, 'red')
        return False

    return f.read()

def getStockDayData(stockcode):
    daydata = []
    dayfilename = tdxdaybase + '/{}.day'.format(stockcode)

    content = file_get_contents(dayfilename, 'rb')
    if(not content):
        return False

    for i in range(0, len(content), 32):
        d = content[i:i+32]
        data = struct.unpack('iiiiifii', d)
        row ={}
        row['date'] = '{}'.format(data[0])
        row['open'] = data[1]/100
        row['high'] = data[2]/100
        row['low'] = data[3]/100
        row['close'] = data[4]/100
        row['amount'] = data[5]/1000
        row['volume'] = data[6]
        daydata.append(row) 
        
    return daydata


def calculateMaData(stockdata, paramlist = [5, 10, 20, 60]):
    closedata = []
    for d in stockdata:
        closedata.append(d['close'])

    madata = {}
    for param in paramlist:
        madata[param] = []
        cnt = 0
        totalvalue = None
        for data in closedata:
            if(cnt < param - 1):
                madata[param].append(None)
            else: #can calculate
                if(not totalvalue):
                    totalvalue = sum(closedata[:param])
                else:
                    totalvalue = totalvalue + closedata[cnt] - closedata[cnt-param]
            
                madata[param].append(totalvalue/param)

            cnt +=1

    return madata

def calculateKdData(stockdata):
    closedata = []
    highdata = []
    lowdata = []
    for d in stockdata:
        closedata.append(d['close'])
        highdata.append(d['high'])
        lowdata.append(d['low'])

    kddata = {}
    kddata['K'] = []
    kddata['D'] = []
        
    #step 1 计算rsv
    cnt = 0
    rsvvals = []
    for data in closedata:
        if(cnt < 9 - 1):
            rsvvals.append(None)
        else: #can calculate
            highvals = highdata[cnt+1-9:cnt+1]
            lowvals = lowdata[cnt+1-9:cnt+1]
            lowval = min(lowvals)
            highval = max(highvals)
            rev = 100
            if(highval-lowval != 0):
                rsv = 100*(data-lowval)/(highval-lowval)
            rsvvals.append(rsv)

        cnt +=1

    #计算k&d
    cnt = 0
    prev_k = 50
    prev_d = 50
    for rsv in rsvvals:
        if(rsv == None):
            kddata['K'].append(None)
            kddata['D'].append(None)
        else:
            k = prev_k*2/3 + rsv/3
            d = prev_d*2/3 + k/3
            kddata['K'].append(k)
            kddata['D'].append(d)
            prev_k = k
            prev_d = d

        cnt +=1        

    return kddata



def calculateMacdData(stockdata):
    closedata = []
    for d in stockdata:
        closedata.append(d['close'])

    dea_param = 9
    short_ema = 12
    long_ema = 26
    lines = ['DIF', 'DEA', 'BAR']
    macddata = {}
    for line in lines:
        macddata[line] = []
        
    #step 1 ema short
    cnt = 0
    emashortvals = []
    for data in closedata:
        if(cnt < short_ema - 1):
            emashortvals.append(None)
        elif(cnt == short_ema - 1):  #day short_ema
            closeval = sum(closedata[:cnt+1])
            dival = closeval/short_ema
            emashortvals.append(dival)
        else: #day after short_ema
            di = closedata[cnt]
            val = emashortvals[cnt-1]*(short_ema-1)/(short_ema+1)+di*2/(short_ema+1)
            emashortvals.append(val)

        cnt +=1

    #step 2 ema long
    cnt = 0
    emalongvals = []
    for data in closedata:
        if(cnt < long_ema - 1):
            emalongvals.append(None)
        elif(cnt == long_ema - 1):  #day long_ema
            closeval = sum(closedata[:cnt+1])
            dival = closeval/long_ema
            emalongvals.append(dival)
        else: #day after long_ema
            di = closedata[cnt]
            val = emalongvals[cnt-1]*(long_ema-1)/(long_ema+1)+di*2/(long_ema+1)
            emalongvals.append(val)

        cnt +=1


    #计算dif
    cnt = 0
    for emalong in emalongvals:
        if(emalong == None):
            macddata['DIF'].append(None)
        else:
            dif =  emashortvals[cnt] - emalong
            macddata['DIF'].append(dif)

        cnt +=1

    #计算DEA
    cnt = 0
    difcnt = 0
    for dif in macddata['DIF']:
        if(dif == None):
            macddata['DEA'].append(None)
        elif(difcnt< dea_param - 1):
            macddata['DEA'].append(None)
            difcnt += 1
        elif(difcnt == dea_param - 1):
            dea = sum(macddata['DIF'][cnt-dea_param+1:cnt+1])/dea_param
            macddata['DEA'].append(dea)
            difcnt += 1
        else:
            dea = macddata['DEA'][cnt-1]*(dea_param-1)/(dea_param+1)+macddata['DIF'][cnt]*2/(dea_param+1)
            macddata['DEA'].append(dea)

        cnt +=1

    #计算bar
    cnt = 0
    difcnt = 0
    for dea in macddata['DEA']:
        if(dea == None):
            macddata['BAR'].append(None)
        else:
            macd = 2*(macddata['DIF'][cnt] - dea)
            macddata['BAR'].append(macd)

        cnt +=1

    return macddata


data = getStockDayData(stockcode)
madata = calculateMaData(data)
kddata = calculateKdData(data)
macddata = calculateMacdData(data)

from sklearn import preprocessing

ylist = []
guessdate = []
result = []
nextday = 1
y = []
x = []
scaler = None
svc = None
from sklearn import svm

for i in range(61, len(data)-nextday):

    next = i + nextday
    next_close = data[next]['close']
    close = data[i]['close']
    if(next_close>close) :
        y.append(1)
    else:
        y.append(-1)

    x.append([
        data[i]['close'],
        #data[i]['amount'],
        madata[5][i], 
        madata[10][i],
        madata[20][i], 
        madata[60][i], 
        #kddata['K'][i], 
        #kddata['D'][i],
        #macddata['DEA'][i],
        #macddata['DIF'][i],
        #macddata['BAR'][i],
        ])

scaler = preprocessing.StandardScaler(copy=True, with_mean=False, with_std=True).fit(x)
X_scaled = scaler.transform(x)
#print(X_scaled[1:10])
#break
testlen = 300
totallen = len(data)
#print(totallen)
#break
tranlen = totallen - testlen
x_tran = X_scaled[0:tranlen]
y_tran = y[0:tranlen]
x_test = X_scaled[tranlen:totallen]
y_test = y[tranlen:totallen]
guessdate = data[tranlen:totallen]


svc = svm.SVC(kernel='poly', degree=3)
svc.fit(x_tran, y_tran)

correct = 0
lastguess = 0
for i in range(0, len(y_test)):
 y_guess = svc.predict(x_test[i])
 lastguess = y_guess
 ylist.append(y_guess[0])
 if(y_guess == y_test[i]):
    correct = correct + 1

result.append([nextday, correct/len(y_test),lastguess,])
#print(nextday)



import pprint
pprint.pprint(result)


for i in range(0, len(ylist)):
    print("step {} @{}: real: {},  guess: {}".format(i, guessdate[i+61]['date'], y_test[i], ylist[i]))
#print(ylist)
#print(y_test)


if(0) :
    for i in range(len(data)-nextday, len(data)):
        x = []
        x.append([
            data[i]['close'],
            madata[5][i], 
            madata[10][i],
            madata[20][i], 
            madata[60][i], 
            ])    

        X_scaled = scaler.transform(x)
        print(svc.predict(X_scaled[0]))
