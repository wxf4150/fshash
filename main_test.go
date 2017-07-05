package main
import(
"testing"
"github.com/bmizerany/assert"
	"fmt"
	"os/exec"
	"encoding/hex"
	"crypto/sha1"
)

//整个处理用过程用到的大多数以函数所且到的数据，都是通过chan 从其它进程获得，单函数做测试不太可行．
//所以测试，主要做个宏观的测试：　新建三五个文件，运行程序看，看得到的每个文件输出的hash值是不是符合预期．

func Test_main(t *testing.T) {
	file1:="/tmp/fshash/a/b/c/file1.txt"
	file2:="/tmp/fshash/a1/b1/c1/file1.txt"
	shell:= fmt.Sprintf(`
	mkdir -p $(dirname %s)
	echo 1>%s
	mkdir -p $(dirname %s)
	echo 2>%s
	`,file1,file1,file2,file2)
	t.Log(shell)
	//生成file1　file２
	err:=exec.Command("sh","-c",shell).Run()
	assert.Equal(t,nil,err,"shell　应该正常运行")

	//wxf@dell:~/go/src/fshash$ echo 1 > /tmp/test.txt
	//wxf@dell:~/go/src/fshash$ du -sh /tmp/test.txt
	//4.0K	/tmp/test.txt
	//虽然文件只有一个字符，在磁盘是也占4k空间．

	f1_sha1:=sha1.New()
	bscount,_:=f1_sha1.Write([]byte("1"))
	exp_file1Hash:=fmt.Sprintf("%s,%s,%d",file1,hex.EncodeToString(f1_sha1.Sum(nil)),bscount)

	f1_sha1.Reset()
	bscount,_=f1_sha1.Write([]byte("2"))
	exp_file2Hash:=fmt.Sprintf("%s,%s,%d",file2,hex.EncodeToString(f1_sha1.Sum(nil)),bscount)

	t.Log(exp_file1Hash)
	t.Log(exp_file2Hash)

}
