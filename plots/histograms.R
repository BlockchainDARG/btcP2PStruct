require(ggplot2)

plot <- function(data) {
  edata <- data.frame()
  prev_tp <- 0
  prev_fp <- 0
  id <- 0
  for(i in 1:nrow(data)) {
    row <- data[i,]
    # do stuff with row
    peers <- row$tp - prev_tp
    if (peers > 0) {
      for(j in 1:peers) {
        edata <- rbind(edata, c(id, F, row$threshold-10))
        id <- id + 1
      }
    }
    prev_tp <- row$tp
    knownNodes <- row$fp - prev_fp
    
    if (knownNodes > 0) {
      for(j in 1:knownNodes) {
        edata <- rbind(edata, c(id, T, row$threshold-10))
        id <- id + 1
      }
    }
    prev_fp <- row$fp
  }
  colnames(edata)  <- c("id", "isNotPeer", "time")
  edata$isNotPeer <- factor(edata$isNotPeer)
  number_ticks <- function(n) {function(limits) pretty(limits, n)}
  g <- guide_legend(title="is peer?")
  
  m <- ggplot(edata, aes(x=edata$time, fill=edata$isNotPeer))
  m + geom_bar(binwidth=10) + geom_bar(binwidth=10, colour="#666666", show_guide=FALSE) +
     coord_cartesian(ylim=c(0,20)) + 
     xlab("now - timestamp (in minutes)") +
    ggtitle("Typical getaddr timestamps") +
   scale_fill_discrete(name="Is peer?", labels=c("T", "F")) +
   scale_x_continuous(breaks=number_ticks(8), limits=c(0, 140))+ 
    theme(axis.text = element_text(size = 25))  + 
  theme(axis.title = element_text(size = 25))  + 
  theme(text = element_text(size = 25))  
}

data <- read.csv("../data/full_node/2015-01-14_15:03:28_time.csv")
png('full_node-histogram.png', width=1024)
plot(data)
dev.off()

data <- read.csv("../data/client/2015-01-14_15:50:51_time.csv")
png('client-histogram.png')
plot(data)
dev.off()
