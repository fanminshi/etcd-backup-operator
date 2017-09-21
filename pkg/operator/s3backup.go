package operator

// func (b *Backup) saveSnap(clusterName string) (int64, error) {
// 	podList, err := b.kubecli.Core().Pods(b.namespace).List(k8sutil.ClusterListOpt(clusterName))

// 	var pods []*v1.Pod
// 	for i := range podList.Items {
// 		pod := &podList.Items[i]
// 		if pod.Status.Phase == v1.PodRunning {
// 			pods = append(pods, pod)
// 		}
// 	}

// 	if len(pods) == 0 {
// 		msg := "no running etcd pods found"
// 		logrus.Warning(msg)
// 		return lastSnapRev, fmt.Errorf(msg)
// 	}
// 	member, rev := getMemberWithMaxRev(pods, nil)
// 	if member == nil {
// 		logrus.Warning("no reachable member")
// 		return lastSnapRev, fmt.Errorf("no reachable member")
// 	}

// 	log.Printf("saving backup for cluster (%s)", clusterName)
// 	if err := b.writeSnap(member, rev); err != nil {
// 		err = fmt.Errorf("write snapshot failed: %v", err)
// 		return lastSnapRev, err
// 	}
// 	return rev, nil
// }

// func (b *Backup) writeSnap(m *etcdutil.Member, rev int64) error {
// 	cfg := clientv3.Config{
// 		Endpoints:   []string{m.ClientURL()},
// 		DialTimeout: constants.DefaultDialTimeout,
// 		TLS:         nil,
// 	}
// 	etcdcli, err := clientv3.New(cfg)
// 	if err != nil {
// 		return fmt.Errorf("failed to create etcd client (%v)", err)
// 	}
// 	defer etcdcli.Close()

// 	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultRequestTimeout)
// 	resp, err := etcdcli.Maintenance.Status(ctx, m.ClientURL())
// 	cancel()
// 	if err != nil {
// 		return err
// 	}

// 	ctx, cancel = context.WithTimeout(context.Background(), constants.DefaultSnapshotTimeout)
// 	rc, err := etcdcli.Maintenance.Snapshot(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to receive snapshot (%v)", err)
// 	}
// 	defer cancel()
// 	defer rc.Close()

// 	_, err = b.be.save(resp.Version, rev, rc)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func getMemberWithMaxRev(pods []*v1.Pod, tc *tls.Config) (*etcdutil.Member, int64) {
// 	var member *etcdutil.Member
// 	maxRev := int64(0)
// 	for _, pod := range pods {
// 		m := &etcdutil.Member{
// 			Name:         pod.Name,
// 			Namespace:    pod.Namespace,
// 			SecureClient: tc != nil,
// 		}
// 		cfg := clientv3.Config{
// 			Endpoints:   []string{m.ClientURL()},
// 			DialTimeout: constants.DefaultDialTimeout,
// 			TLS:         tc,
// 		}
// 		etcdcli, err := clientv3.New(cfg)
// 		if err != nil {
// 			logrus.Warningf("failed to create etcd client for pod (%v): %v", pod.Name, err)
// 			continue
// 		}
// 		defer etcdcli.Close()

// 		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultRequestTimeout)
// 		resp, err := etcdcli.Get(ctx, "/", clientv3.WithSerializable())
// 		cancel()
// 		if err != nil {
// 			logrus.Warningf("getMaxRev: failed to get revision from member %s (%s)", m.Name, m.ClientURL())
// 			continue
// 		}

// 		logrus.Infof("getMaxRev: member %s revision (%d)", m.Name, resp.Header.Revision)
// 		if resp.Header.Revision > maxRev {
// 			maxRev = resp.Header.Revision
// 			member = m
// 		}
// 	}
// 	return member, maxRev
// }
