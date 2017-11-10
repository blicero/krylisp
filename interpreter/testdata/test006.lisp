;; /home/krylon/go/src/krylisp/interpreter/testdata/test006.lisp
;; created on 07. 11. 2017
;; (c) 2017 Benjamin Walkenhorst
;; Time-stamp: <2017-11-10 20:35:41 krylon>
;;
;; Dienstag, 07. 11. 2017, 21:27
;; I think it is about time to add I/O facilities to kryLisp.
;; I would like to follow Go's example and present a unified interface for I/O
;; for both file I/O and network I/O.
;; But creating such an interface is not going to be terribly easy.
;; So I created this test script as a playground. Once I get something I am
;; happy with, I can implement the needed parts in my interpreter.

(define path (concat (getenv "HOME") "test.file"))

path


