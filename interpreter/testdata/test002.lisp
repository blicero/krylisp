;; test002.lisp
;; Time-stamp: <2017-10-31 00:24:02 krylon>
;;
;; Okay, on to the second test script! \o/


(defun factorial (x) 
  (if (< x 2) 1
    (* x (factorial (- x 1)))))


(define num 0)

(set! num (factorial 10))


