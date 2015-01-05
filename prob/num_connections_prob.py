import scipy.stats
from scipy.stats import binom

def binofit_scalar(x,n,alpha):
    '''Parameter estimates and confidence intervals for binomial data.
    (p,ci) = binofit(x,N,alpha)

    Source: Matlab's binofit.m
    Reference:
      [1]  Johnson, Norman L., Kotz, Samuel, & Kemp, Adrienne W.,
      "Univariate Discrete Distributions, Second Edition", Wiley
      1992 p. 124-130.
    http://books.google.com/books?id=JchiadWLnykC&printsec=frontcover&dq=Univariate+Discrete+Distributions#PPA131,M1

    Re-written by Santiago Jaramillo - 2008.10.06
    '''
    
    if n<1:
        Psuccess = np.NaN
        ConfIntervals = (np.NaN,np.NaN)
    else:
        Psuccess = float(x)/n
        nu1 = 2*x
        nu2 = 2*(n-x+1);
        F = scipy.stats.f.ppf(alpha/2,nu1,nu2)
        lb  = (nu1*F)/(nu2 + nu1*F)
        if x==0: lb=0
        nu1 = 2*(x+1)
        nu2 = 2*(n-x)
        F = scipy.stats.f.ppf(1-alpha/2,nu1,nu2)
        ub = (nu1*F)/(nu2 + nu1*F)
        if x==n: ub=1
        ConfIntervals = (lb,ub);
    return (Psuccess,ConfIntervals)

# n: number of messages
n = 5*625
# o: number of attacker connections
o = 55
assert(o >=2)
# c: number of non-attacker connections
c = 59

# t: number of total connections
# h: number of positive outcomes (receive addr from one of attacker connections) if no "random noise"
def toP(c,o):
    t = c + o 
    return (float(o-1)/t)

h = int(toP(c,o)*n)
alpha = 0.01

p, pci = binofit_scalar(h, n, alpha)

# compute number of non-attacker connections from success probability
def toC(p):
    return ((-1)+(1-p)*o)/p
print "%d%% conf interval" % (100 * (1-alpha)), toC(pci[1]), "<=", toC(p), "<=", toC(pci[0])


# doesn't seem possible to estimate this kind of error
# compute for every non-attacker prob that exactly this outcome
# of h takes place
sum = 0
r = 5
p_true = toP(c,o)
for c_est in range(100):
    # print c_est, binom.pmf(h, n, toP(c_est,o))
    if abs(c-c_est) > 3:
        p_est = toP(c_est,o)
        sum += binom.pmf(h, n, p_est)
print "sum of prob for wrong c estimate:", sum, ",sum of prob for true c estimate:", binom.pmf(h, n, p_true)
