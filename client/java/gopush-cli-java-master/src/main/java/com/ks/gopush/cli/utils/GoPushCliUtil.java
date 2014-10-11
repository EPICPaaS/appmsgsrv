package com.ks.gopush.cli.utils;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;

import org.junit.Assert;

import com.ks.gopush.cli.GoPushCli;
import com.ks.gopush.cli.Listener;
import com.ks.gopush.cli.PushMessage;
import com.ks.gopush.cli.model.request.BaseRequest;
import com.ks.gopush.cli.model.request.MsgRequest;
import com.ks.gopush.cli.model.response.MsgResponse;


/**
 * gopush java客户端工具类
 * @author ghg
   */
public class GoPushCliUtil {
	
	/**
	  *    应用发送消息给用户
	 * @param host   应用消息IP地址
	 * @param port   应用消息服务端口
	 * @param toKen  应用的私有Token
	 * @param msgReq 消息请求对象
	 * @return 消息返回体  ret = 0标示成功
	 * @throws IOException 抛出异常
	 */
	public static MsgResponse appSendMsgToUser(String host ,Integer port,MsgRequest msgReq) throws IOException{
		String data = JsonMapper.buildNormalMapper().toJson(msgReq);
		String url = HttpUtils.getURL("http", host, port,Constant.APP_CLIENT_APP_USER_PUSH);
		String ret = 	HttpUtils.post(url, data);
		MsgResponse msgresp =	JsonMapper.buildNormalMapper().fromJson(ret, MsgResponse.class);
		return msgresp;
	}

	
	
	public static void main(String[] args) {
		
		MsgRequest msgr = new MsgRequest();
		msgr.setContent("ssssssssssssssss");
		msgr.setMsgType("1");
		
		BaseRequest	tmp = new BaseRequest();
		tmp.setToken("eflow_token");
		msgr.setBaseRequest(tmp);
		
		String[] toUserNames = {"23622391649370234@user","22622391649384527@user"};
		msgr.setToUserNames(toUserNames);
		
		Map<String,String> objectContent = new HashMap<String, String>();
		objectContent.put("appId", "2222");
		objectContent.put("appSendId", "111");
		msgr.setObjectContent(objectContent);

		
		String host = "10.180.120.63";
		Integer port = 8093;
		
		try {
		MsgResponse msgresp =	appSendMsgToUser(host, port, msgr);
		System.out.println(msgresp.getBaseResponse().getRet());
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		
		String urlPath = "http://10.180.120.63:8090/1/msg/get?k=23622391649370234@user&m=0";
		try {
			String result =		HttpUtils.get(urlPath);
			System.out.println(result);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		
	}
	
}
