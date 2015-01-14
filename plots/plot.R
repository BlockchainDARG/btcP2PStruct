data <- read.csv("../data/full_node/2015-01-14_15:03:28_time.csv")
data$tpr <- data$tp/(data$tp+data$fn)

png('plot.png')
plot(data$fp, data$tpr, xlim=c(0,10), 
     ylim=c(0.05, 0.25), 
     main="ROC for strategy with varying threshold",
     xlab = "false positives",
     ylab = "true positive rate")
dev.off()