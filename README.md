`gogit` could send the build status to different git providers. Such as:

* GitHub
* Gitlab (public or private)

## Usage
Below is an example of sending build status to a private Gitlab server:

```shell
gogit --provider gitlab \
  --server http://10.121.218.82:6080 \
  --repo yaml-readme \
  --pr 1 \
  --username linuxsuren \
  --token h-zez9CWzyzykbLoS53s
```

## TODO
* Support more git providers

## Thanks
Thanks to these open source projects, they did a lot of important work.
* github.com/jenkins-x/go-scm
* github.com/spf13/cobra
