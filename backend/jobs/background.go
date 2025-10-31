package jobs

import (
	arp "goaway/backend/dns"
	"goaway/backend/logging"
	"goaway/backend/services"
)

var log = logging.GetLogger()

type BackgroundJobs struct {
	registry *services.ServiceRegistry
}

func NewBackgroundJobs(registry *services.ServiceRegistry) *BackgroundJobs {
	return &BackgroundJobs{
		registry: registry,
	}
}

func (b *BackgroundJobs) Start(readyChan <-chan struct{}) {
	b.startHostnameCachePopulation()
	b.cleanVendorResponseCache(readyChan)
	b.startARPProcessing(readyChan)
	b.startScheduledUpdates(readyChan)
	b.startCacheCleanup(readyChan)
	b.startPrefetcher(readyChan)
}

func (b *BackgroundJobs) startHostnameCachePopulation() {
	if err := b.registry.Context.DNSServer.PopulateHostnameCache(); err != nil {
		log.Warning("Unable to populate hostname cache: %s", err)
	}
}

func (b *BackgroundJobs) startARPProcessing(readyChan <-chan struct{}) {
	go func() {
		<-readyChan
		log.Debug("Starting ARP table processing...")
		arp.ProcessARPTable()
	}()
}

func (b *BackgroundJobs) cleanVendorResponseCache(readyChan <-chan struct{}) {
	go func() {
		<-readyChan
		log.Debug("Starting vendor response table processing...")
		arp.CleanVendorResponseCache()
	}()
}

func (b *BackgroundJobs) startScheduledUpdates(readyChan <-chan struct{}) {
	go func() {
		<-readyChan
		if b.registry.Context.Config.Misc.ScheduledBlacklistUpdates {
			log.Debug("Starting scheduler for automatic list updates...")
			b.registry.BlacklistService.ScheduleAutomaticListUpdates()
		}
	}()
}

func (b *BackgroundJobs) startCacheCleanup(readyChan <-chan struct{}) {
	go func() {
		<-readyChan
		log.Debug("Starting cache cleanup routine...")
		b.registry.Context.DNSServer.ClearOldEntries()
	}()
}

func (b *BackgroundJobs) startPrefetcher(readyChan <-chan struct{}) {
	go func() {
		<-readyChan
		log.Debug("Starting prefetcher...")
		b.registry.PrefetchService.Run()
	}()
}
