;; test001.lisp
;; Time-stamp: <2017-10-15 14:16:37 krylon>
;;
;; I am going to try to implement a few primitive Lisp operations in pure
;; Lisp to see how that works out.
;; I will probably replace with them with special forms implemented in Go,
;; if only for performance reasons, but I want to try, for a moment,
;; how far I could push this implementing-Lisp-in-Lisp-thing.

(defun map (f lst)
  "Return a list of the results of applying f to each element of lst."
  (if (nil? lst) nil
    (cons (apply f (list (car lst))) ;(apply f (list (car lst)))
	  (map f (cdr lst)))))

(define numbers (list 1 2 3))

(defun twice (x)
  "Returns twice its input."
  (* x 2))

(map #twice numbers)


