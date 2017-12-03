;; /home/krylon/go/src/krylisp/interpreter/testdata/test007.lisp
;; created on 13. 11. 2017
;; (c) 2017 Benjamin Walkenhorst
;; Time-stamp: <2017-11-25 15:10:41 krylon>

(define fh (fopen filename :direction :read :permission 0644))

(define acc 0)

(defun num-or-zero (x)
  (if (nil? x) 0 x))

(while (not (feof fh))
  (let ((line (fread-line fh)))
    (print line)
    (set! acc (+ acc (num-or-zero (read-from-string line))))))

acc

