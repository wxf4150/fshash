package cfg
import (
"github.com/vharitonsky/iniflags"
"flag"
)

var(
Filter= flag.String("filter", "", "需要忽略的文件或目录,支持操作系统文件查找所用的通配符*?[a-z]等,默认不忽略任何文件 例如：*.php会忽略所有的php文件; /tmp/wxf/w* 会忽略扆/tmp/wxf/文件夹下w开头的文件夹")
RootDir= flag.String("root", "/usr/local", "需要处理的文件夹.默认/usr/local目录.　对于大目录，第二次运行时，因为文件系统的缓存，用时明显比第一次缩小很多")
Out= flag.String("out", "./sha1.txt", "结果输出文件")
)
func init() {
	iniflags.Parse();
}
