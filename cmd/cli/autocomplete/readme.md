
## 安装

```
$ sudo install turingchain-cli.bash  /usr/share/bash-completion/completions/turingchain-cli
```

不安装
```
. turingchain-cli.bash
```

```
# 重新开个窗口就有用了
$ ./turingchain/turingchain-cli 
account   trc       coins     evm       hashlock  mempool   privacy   retrieve  send      ticket    trade     version   
block     close     config    exec      help      net       relay     seed      stat      token     tx        wallet    
```

## 演示
```
linj@linj-TM1701:~$ ./turingchain/turingchain-cli 
account   trc       coins     evm       hashlock  mempool   privacy   retrieve  send      ticket    trade     version   
block     close     config    exec      help      net       relay     seed      stat      token     tx        wallet    
linj@linj-TM1701:~$ ./turingchain/turingchain-cli b
block  trc    
linj@linj-TM1701:~$ ./turingchain/turingchain-cli trc 
priv2priv  priv2pub   pub2priv   send       transfer   txgroup    withdraw   
linj@linj-TM1701:~$ ./turingchain/turingchain-cli trc t
transfer  txgroup   
linj@linj-TM1701:~$ ./turingchain/turingchain-cli trc transfer -
-a        --amount  -h        --help    -n        --note    --para    --rpc     -t        --to  
```


## 原理

 地址 https://confluence.__officeSite__/pages/viewpage.action?pageId=5967798
