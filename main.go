package main

import (
	"os"
	"runtime"
	"path/filepath"
	"sync"
	"fmt"
	"fshash/cfg"
	"time"
	"log"
	"crypto/sha1"
	"io"
	"bufio"
	"encoding/hex"
)

var taskWg sync.WaitGroup
var recordWg sync.WaitGroup
func main(){
	log.Printf("开始处理文件夹%s  filter:%s",*cfg.RootDir,*cfg.Filter)
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)

	stime:=time.Now()

	go func(){
		taskWg.Add(1)
		err:=filepath.Walk(*cfg.RootDir, walkFunc)
		if err!=nil{
			log.Println(err)
		}
		close(visitDataChan)
		taskWg.Add(-1)
	}()
	for i := 0; i < numCpu; i++ {
		taskWg.Add(1)
		go taskItem()
	}
	go recordHash(*cfg.Out)

	taskWg.Wait()
	//hash处理现成后关闭，关闭中转hashvalue的chan
	close(hashValuesChan)
	recordWg.Wait()//阻塞直到recordHash执行完成．

	log.Printf("文件夹处理完成,花费时间%f秒",time.Since(stime).Seconds())

}

//walkFunc给管道发送的数据结构（path fileinfo）
type visitData struct {
	path string
	info os.FileInfo
}
var visitDataChan chan *visitData =make(chan *visitData,3000)
//遍历root下所有文件路径
func walkFunc(fpath string, info os.FileInfo, err error) error {
	//log.Println("walk:",fpath)
	if *cfg.Filter!=""{
		//skipdir
		ok, err1 := filepath.Match(*cfg.Filter, filepath.Dir(fpath));err=err1
		if ok {
			log.Println("skip dir:",filepath.Dir(fpath), info.Name())
			// Filter匹配时，跳过当前目录的包含子目录　和　当前目录的后续文件
			// 注意会跳过子目录
			return filepath.SkipDir
		}
		//skipfile
		ok, err1 = filepath.Match(*cfg.Filter, info.Name());err=err1
		if ok {
			log.Println("skip file:",fpath)
			return nil
		}
	}
	if err==nil{
		taskWg.Add(1)
		visitDataChan <-&visitData{path:fpath,info:info}
	}
	return err
}


func taskItem(){
	for vdata := range visitDataChan {
		//排除链接符号/文件夹　等文件：　ModeDir | ModeSymlink | ModeNamedPipe | ModeSocket | ModeDevice
		if vdata.info.Mode().IsRegular() && vdata.info.Size()>0{
			sh1:=sha1.New()
			filein, err := os.Open(vdata.path);
			if err != nil {
				log.Println("文件打开错误",err, vdata.path)
				continue
			}
			blen,err:=io.Copy(sh1,filein)
			filein.Close()
			if err==nil{
				//line:=fmt.Sprintf("%s %s %d\n",vdata.path,hex.EncodeToString(sh1.Sum(nil)),blen)
				hashValuesChan<-&hashData{path:vdata.path,sum:hex.EncodeToString(sh1.Sum(nil)),size:blen}
				//log.Printf(line)
			}else{
				log.Println("err",err,vdata.path)
			}
		}
		taskWg.Add(-1)
	}
	taskWg.Add(-1)
}
//文件hashSum完的结果 管道，接收字符串格式为：｛文件名,hashSum,Filesize｝
var hashValuesChan chan *hashData=make(chan *hashData,2000)
type hashData struct {
	path string
	sum string
	size int64
}

//不断把hashValuesChan管道的内容写到文件
func recordHash(outFile string){
	var fSizeCounter int64
	var fCount int64
	out,err:=os.Create(outFile);defer out.Close()
	if err!=nil{
		log.Printf("outFile err",err)
		os.Exit(1)
	}
	//bufio writer 默认可以每４k写一次文件，可以用NewWriterSize设置buffer_size
	fwriter:=bufio.NewWriter(out)
	recordWg.Add(1)
	for hdata:=range hashValuesChan{
		line:=fmt.Sprintf("%s,%s,%d\n",hdata.path,hdata.sum,hdata.size)
		fwriter.WriteString(line)
		fSizeCounter+=hdata.size
		fCount++
	}
	fwriter.Flush()
	recordWg.Done()
	log.Printf("共处理文件数量%d ,累计大小：%f兆",fCount,float32(fSizeCounter)/1024/1024)
}