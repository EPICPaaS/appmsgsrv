package com.ks.gopush.cli;

import java.util.ArrayList;
import java.util.concurrent.TimeUnit;

import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;

public class GoPushCliTest {
	//@Before
	public void init() {
		local.set(new GoPushCli("127.0.0.1", 8090, "a", 300, 0L, 0L,
				new Listener() {
					@Override
					public void onOpen() {
						System.err.println("dang dang dang dang~open");
					}

					@Override
					public void onOnlineMessage(PushMessage message) {
						System.err.println("online message: "
								+ message.getMsg());
						System.err.println("online message id: "
								+ message.getMid());
					}

					@Override
					public void onOfflineMessage(ArrayList<PushMessage> messages) {
						if (messages != null)
							for (PushMessage message : messages) {
								System.err.println("offline message: "
										+ message.getMsg());
								System.err.println("offline message id: "
										+ message.getMid());
							}
					}

					@Override
					public void onError(Throwable e, String message) {
						Assert.fail(message);
					}

					@Override
					public void onClose() {
						System.err.println("pu pu pu pu~");
					}
				}));
	}
//$  curl -d "{\"test\":1}" http://localhost:8091push/private?key=a\&expire=600

  //@Test
	public void testNoSync() {
		GoPushCli cli = local.get();
		cli.start(false);

		Assert.assertTrue("获取节点失败", cli.isGetNode());
		Assert.assertTrue("握手失败", cli.isHandshake());
		cli.destory();
	}

	//@Test
	public void testSync() {
		final GoPushCli cli = local.get();
		new Thread() {
			public void run() {
				cli.start(true);
			}
		}.start();
		try {
			TimeUnit.SECONDS.sleep(100);
		} catch (InterruptedException e) {
		}
		Assert.assertTrue("获取节点失败", cli.isGetNode());
		Assert.assertTrue("握手失败", cli.isHandshake());
		try {
			TimeUnit.SECONDS.sleep(2000000);
		} catch (InterruptedException e) {
		}
		cli.destory();
	}

	private ThreadLocal<GoPushCli> local = new ThreadLocal<GoPushCli>();
}
