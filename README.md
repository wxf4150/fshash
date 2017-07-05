# files hash

目录所有文件hash 列表． 大概思路如下

## 最简单的方法
伪代码大概的如下：
```
scanDir(dir string){
    fnames:=getFnames(dir)
    for(name in fnames){
        fullpath:= (Path.Join(dir,name))
        if isDir(fullpath){
            scanDir(fullpath)
        }else{
            fmt.println(fullpath,sha1(fullpath),size) 
        }
   
    }
}
```
但明显这是单线程，多核明显浪费

## 多线程
发生递归时用go scanDir
```
runtime.GOMAXPROCS(runtime.NumCPU())
scanDir(dir string){
    fnames:=getFnames(dir)
    for(name in fnames){
        fullpath:= (Path.Join(dir,name))
        if isDir(fullpath){
            go scanDir(fullpath)
        }else{
            fmt.println(fullpath,sha1(fullpath),size) 
        }
   
    }
}

```


## filePath.Walk
想用filePath.Match作过滤时，学习发现有更好filePath.Walk直接返回全部文件更好用．
但是filePath.Walk 又是单线程的．　

看到github 有parallel版的filePath.Walk
https://github.com/MichaelTJones/walk
学习了下相关源代码：并发16线程，每线程处理一个文件夹 ;fileinfo:=os.Lstat(filepath)不处理相关软链接的用法.


## sha１效率
随后又想，即使我的 parallel版的Walk 性能提升4-6倍．　sha1的效率呢．
https://stackoverflow.com/questions/11985238/how-long-does-a-sha1-hash-take-to-generate
4k的文件sha1每秒能处理400－500个． 但walk出一个500个文件列表数据，应该更是极速,文件列表花时间并不多．
所以应该让cpu所有核尽可能的都在处理sha1运算,应该并行处理的应该主要是读文件内容并sha1运算．


## 最终算法:
伪代码
```
main(){

    walkers := runtime.NumCPU()
    runtime.GOMAXPROCS(walkers)
    
    go filepath.Walk(`/root`, walkFunc)
    for i := 0; i < walkers; i++ {
        wg.Add(1)
        go taskItem()
    }
    wg.Wait()

}

var wg sync.WaitGroup
func walkFunc(path string, info os.FileInfo, err error) error {
    
	//ok, err := filepath.Match(filterPatten, info.Name())
	ok,err:=regFilter.math(path)
	if ok {
		fmt.Println("skip:",filepath.Dir(path), info.Name())
		// 遇到 txt 文件则继续处理所在目录的下一个目录
		// 注意会跳过子目录
		return filepath.SkipDir
	}else{
	    wg.Add(1)
	    fnameChan<-fullPath
	 
	 }
	return err
}

//并行处理hash
func taskItem(){
    for file := range fnameChan {
    		fmt.println(file,sha1(file),size) 
    		wg.Add(-1)
    	}
    
}
```


## 更快的hash算法
https://github.com/minio/blake2b-simd



## 运行方式：
```
go get github.com/vharitonsky/iniflags
go get github.com/wxf4150/fshash
cd $GOPATH/src/github.com/wxf4150/fshash
go run main.go --help
```

## 运行性能结果
```
wxf@dell:~/go/src/fshash$ go run main.go
2017/07/05 23:30:32 开始处理文件夹/usr/local  filter:
2017/07/05 23:30:34 共处理文件数量27327 ,累计大小：1932.424194兆
2017/07/05 23:30:34 文件夹处理完成,花费时间1.739108秒
wxf@dell:~/go/src/fshash$ ls
cfg  main.go  readme.md  sha1.txt
wxf@dell:~/go/src/fshash$ ll ./sha1.txt -h
-rw-rw-r-- 1 wxf wxf 2.5M 7月   5 23:30 ./sha1.txt
wxf@dell:~/go/src/fshash$ head ./sha1.txt 
/usr/local/bin/mine,c2c1f697bc0a25714d01ac78eb154d50b879cdb7,2678
/usr/local/bin/charm,0ffbcdde9a851ec579f6b55449defb03d217e10b,2675
/usr/local/bin/watchman-make,ef75d85e4b02b1368d0f33cd76062d50c0a35007,7744
/usr/local/bin/watchman-wait,5de74459b6da5425490acaccc698d5e3fec7f2b1,7793
/usr/local/bin/wstorm,bbe6fea6e79f43beb88ae986a27eaee4e6c80e34,2769
/usr/local/bin/xml2-config,2d8cff6aac7598318d5749ad4f7ec9b6819d9110,1642
/usr/local/bin/xmlcatalog,8ca9a205ae45e71ebe91b6b73264604bee79c00d,47522
/usr/local/bin/xmllint,0f5aae5a5b82c01071c85994c9d472c9f75baa93,231570
/usr/local/go/AUTHORS,30e1c376f21d09b982eebd1072ca5d1605fa2c50,33243
/usr/local/go/CONTRIBUTING.md,a481e0601ce9b6e36d080dfe375fa6f8d66e8ea6,1366

```
