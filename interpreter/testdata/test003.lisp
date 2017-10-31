;; test003.lisp
;; Time-stamp: <2017-10-31 00:37:49 krylon>
;;
;; Same as test002.lisp, except that we use bignums.


(defun factorial (x) 
  (if (< x 2b) 1b
    (* x (factorial (- x 1b)))))


(define num 0b)

(set! num (factorial 10b))


