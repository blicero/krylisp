;; /home/krylon/go/src/krylisp/interpreter/testdata/test007.lisp
;; created on 13. 11. 2017
;; (c) 2017 Benjamin Walkenhorst
;; Time-stamp: <2017-11-13 18:43:09 krylon>

(define fh (fopen filename "r"))

(define acc 0)

(while (not (feof fh))
  (let ((line (fread-line fh)))
    (set! acc (+ acc (read-from-string line)))))


