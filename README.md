# OpenAList
[中文](./README_cn.md) |  English

## Notice
This is a fork of [Alist](https://github.com/alist-org/alist) based on version 3.45.0.

The document site has been deployed: http://alist.iots.vip/

As for obtaining tokens for various cloud drives, it is strongly recommended to use an offline solution (the original API provided by the project's author has been replaced with a black hole to avoid security risks).  


## Description
OpenAList is an forked version of the original Alist file list program.

Considering that several forked organizations are still in the onboarding phase and it's hard to distinguish genuine ones from fake.  

For personal use, I have already forked and modified this part of the code.   

If you need it urgently, you can directly use my image tag  `alliot/alist:latest`.
> Please note, since the static password salt has been changed, you need to reset the password when using this image
> `docker exec -it alist /bin/sh`
> Then execute `./alist admin set my_new_password`

The image was built from the following repository CIs, feel free to review them yourself:  
https://github.com/AlliotTech/openalist  
https://github.com/AlliotTech/openalist-web  
https://github.com/AlliotTech/openalist-docs  


## Features
- Original Alist features
- etc.

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.


## Acknowledgments
- Original [Alist Project](https://github.com/alist-org/alist)

## More  
https://github.com/AlistGo/alist/issues/8649  
https://github.com/AlistGo/alist/issues/8651  
...
