package com.ks.gopush.cli.utils;

public class Constant {
	public static final int KS_NET_STATE_OK = 0;

	public static final int KS_NET_EXCEPTION_SUBSCRIBE_CODE = -1;
	public static final int KS_NET_EXCEPTION_OFFLINE_CODE = -2;
	public static final int KS_NET_EXCEPTION_SOCKET_READ_CODE = -3;
	public static final int KS_NET_EXCEPTION_SOCKET_WRITE_CODE = -4;
	public static final int KS_NET_EXCEPTION_SOCKET_INIT_CODE = -5;

	public static final String KS_NET_JSON_KEY_RET = "ret";
	public static final String KS_NET_JSON_KEY_MSG = "msg";
	public static final String KS_NET_JSON_KEY_DATA = "data";
	public static final String KS_NET_JSON_KEY_SERVER = "server";
	
	public static final String KS_NET_JSON_KEY_MESSAGES = "msgs";
	public static final String KS_NET_JSON_KEY_PMESSAGES = "pmsgs";
	
	public static final String KS_NET_JSON_KEY_MESSAGE_MSG = "msg";
	public static final String KS_NET_JSON_KEY_MESSAGE_MID = "mid";
	public static final String KS_NET_JSON_KEY_MESSAGE_GID = "gid";

	public static final String KS_NET_KEY_ADDRESS = "address";
	public static final String KS_NET_KEY_PORT = "port";

	public static final String KS_NET_SOCKET_CONNECTION_ACTION = "socket_connection_action";

	public static final int KS_NET_MESSAGE_OBTAIN_DATA_OK = 2;
	public static final int KS_NET_MESSAGE_DISCONNECT = 1;
	
	public static final int KS_NET_MESSAGE_PRIVATE_GID = 0;
	
	/**
	 * response code
	 */
	public static final int OK             = 0;
	public static final int NotFoundServer = 1001;
	public static final int NotFound       = 65531;
	public static final int TooLong        = 65532;
	public static final int AuthErr        = 65533;
	public static final int ParamErr       = 65534;
	public static final int InternalErr    = 65535;
	
	
	
	public static final String V1_SERVER_GET = "/1/server/get";
	public static final String V1_MSG_GET = "/1/msg/get";
	public static final String V1_TIME_GET = "/1/time/get";
	
	public static final String SERVER_GET = "/server/get";
	public static final String MSG_GET = "/msg/get";
	public static final String TIME_GET = "/time/get";
	
	public static final String V1_ADMIN_PUSH_PRIVATE = "/1/admin/push/private";
	public static final String V1_ADMIN_MSG_DEL = "/1/admin/msg/del";
	
	public static final String ADMIN_PUSH = "/admin/push";
	public static final String ADMIN_MSG_CLEAN = "/admin/msg/clean";
	
	public static final String APP_STATIC = "/app/static/";
	public static final String APP_CLIENT_DEVICE_LOGIN = "/app/client/device/login";
	public static final String APP_CLIENT_DEVICE_PUSH = "/app/client/device/push";
	public static final String APP_CLIENT_DEVICE_ADDORREMOVECONTACT = "/app/client/device/addOrRemoveContact";
	public static final String APP_CLIENT_DEVICE_GETMEMBER = "/app/client/device/getMember";
	public static final String APP_CLIENT_DEVICE_CHECKUPDATE = "/app/client/device/checkUpdate";
	public static final String APP_CLIENT_DEVICE_GETORGINFO = "/app/client/device/getOrgInfo";
	public static final String APP_CLIENT_DEVICE_GETORGUSERLIST = "/app/client/device/getOrgUserList";
	public static final String APP_CLIENT_DEVICE_SYNCORG = "/app/client/device/syncOrg";
	public static final String APP_CLIENT_DEVICE_SEARCH = "/app/client/device/search";
	public static final String APP_CLIENT_DEVICE_CREATEQUN = "/app/client/device/create-qun";
	public static final String APP_CLIENT_DEVICE_GETQUNMEMBERS = "/app/client/device/getQunMembers";
	public static final String APP_CLIENT_DEVICE_UPDATEQUNTOPIC = "/app/client/device/updateQunTopic";
	public static final String APP_CLIENT_DEVICE_ADDQUNMEMBER = "/app/client/device/addQunMember";
	public static final String APP_CLIENT_DEVICE_DELQUNMEMBER = "/app/client/device/delQunMember";
	
	
	/**
	 * 应用 - 用户 消息推送
	 */
	public static final String APP_CLIENT_APP_USER_PUSH = "app/client/app/user/push";
	
	/**
	 * 二维码
	 */
	public static final String APP_USER_ERWEIMA = "/app/user/erweima";
}
